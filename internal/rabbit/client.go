package rabbit

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn       *amqp.Connection
	channel    *amqp.Channel
	config     Config
	mutex      sync.Mutex
	connected  atomic.Bool
	generation atomic.Uint64
	logger     *slog.Logger
}

func NewClient(config Config, logger *slog.Logger) (*Client, error) {
	c := &Client{config: config, logger: logger}
	if err := c.connect(); err != nil {
		return nil, err
	}

	c.logger.Info("rabbit topology ready",
		"exchange", ExchangeNotifications,
		"queues", []string{QueueIntruder, QueueUnknownPerson},
	)
	return c, nil
}

// IsConnected indica si la conexión y el canal con RabbitMQ están activos.
func (c *Client) IsConnected() bool {
	return c.connected.Load()
}

func (c *Client) connect() error {
	conn, err := amqp.Dial(*c.config.URL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	if err := ch.Qos(*c.config.PrefetchCount, 0, false); err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	if err := declareTopology(ch); err != nil {
		ch.Close()
		conn.Close()
		return err
	}

	// La generación evita que el cierre de la conexión anterior marque esta como caída.
	generation := c.generation.Add(1)

	c.mutex.Lock()
	oldChannel := c.channel
	oldConn := c.conn
	c.conn = conn
	c.channel = ch
	c.mutex.Unlock()

	if oldChannel != nil {
		_ = oldChannel.Close()
	}
	if oldConn != nil {
		_ = oldConn.Close()
	}

	// Se marca como desconectado apenas se cae la conexión o el canal, sin esperar
	// a que el consumer lo detecte.
	connClosed := conn.NotifyClose(make(chan *amqp.Error, 1))
	chClosed := ch.NotifyClose(make(chan *amqp.Error, 1))
	c.connected.Store(true)
	go func() {
		select {
		case <-connClosed:
		case <-chClosed:
		}
		if c.generation.Load() == generation {
			c.connected.Store(false)
		}
	}()

	return nil
}

func declareTopology(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(
		ExchangeNotifications,
		amqp.ExchangeTopic,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("declare exchange %s: %w", ExchangeNotifications, err)
	}

	for _, queue := range []string{QueueIntruder, QueueUnknownPerson} {
		if _, err := ch.QueueDeclare(queue, true, false, false, false, nil); err != nil {
			return fmt.Errorf("declare queue %s: %w", queue, err)
		}
		if err := ch.QueueBind(queue, queue, ExchangeNotifications, false, nil); err != nil {
			return fmt.Errorf("bind queue %s: %w", queue, err)
		}
	}

	return nil
}

func (c *Client) reconnect(ctx context.Context) error {
	c.connected.Store(false)
	backoff := 2 * time.Second
	for {
		c.logger.Info("reconnecting...", "backoff", backoff)
		select {
		case <-ctx.Done():
			return fmt.Errorf("reconnect cancelled")
		case <-time.After(backoff):
		}

		if err := c.connect(); err != nil {
			c.logger.Error("reconnect failed", "error", err)
			backoff = min(backoff*2, 60*time.Second)
			continue
		}

		c.logger.Info("reconnected successfully")
		return nil
	}
}

// Consume entrega los mensajes de las dos colas hasta que se cancela el context.
// Al cancelarse, deja de consumir (cancela los consumers en el broker) y cierra el canal devuelto.
func (c *Client) Consume(ctx context.Context) <-chan Message {
	out := make(chan Message)

	go func() {
		defer close(out)
		for {
			if ctx.Err() != nil {
				return
			}

			intruder, unknown, err := c.startConsuming()
			if err != nil {
				c.logger.Error("failed to start consuming", "error", err)
				if err := c.reconnect(ctx); err != nil {
					return
				}
				continue
			}

			if stopped := c.forward(ctx, intruder, unknown, out); stopped {
				return
			}

			c.logger.Warn("connection lost, reconnecting...")
			if err := c.reconnect(ctx); err != nil {
				return
			}
		}
	}()

	return out
}

// forward reenvía las deliveries a out. Devuelve true si se detuvo por cancelación del
// context y false si algún canal de deliveries se cerró (conexión caída).
func (c *Client) forward(ctx context.Context, intruder <-chan amqp.Delivery, unknown <-chan amqp.Delivery, out chan<- Message) bool {
	for {
		select {
		case <-ctx.Done():
			c.stopConsuming()
			return true
		case msg, ok := <-intruder:
			if !ok {
				return false
			}
			if c.deliver(ctx, out, QueueIntruder, msg) {
				return true
			}
		case msg, ok := <-unknown:
			if !ok {
				return false
			}
			if c.deliver(ctx, out, QueueUnknownPerson, msg) {
				return true
			}
		}
	}
}

// deliver encola el mensaje para procesarlo. Devuelve true si hay que detenerse porque
// se canceló el context; en ese caso el mensaje vuelve a la cola.
func (c *Client) deliver(ctx context.Context, out chan<- Message, queue string, msg amqp.Delivery) bool {
	select {
	case out <- Message{Queue: queue, Body: msg.Body, delivery: msg}:
		return false
	case <-ctx.Done():
		if err := msg.Nack(false, true); err != nil {
			c.logger.Error("failed to requeue message on shutdown", "queue", queue, "error", err)
		}
		c.stopConsuming()
		return true
	}
}

func (c *Client) startConsuming() (<-chan amqp.Delivery, <-chan amqp.Delivery, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	intruder, err := c.channel.Consume(QueueIntruder, QueueIntruder, false, false, false, false, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("consume %s: %w", QueueIntruder, err)
	}
	unknown, err := c.channel.Consume(QueueUnknownPerson, QueueUnknownPerson, false, false, false, false, nil)
	if err != nil {
		if cancelErr := c.channel.Cancel(QueueIntruder, false); cancelErr != nil && !errors.Is(cancelErr, amqp.ErrClosed) {
			c.logger.Error("failed to cancel consumer", "consumer", QueueIntruder, "error", cancelErr)
		}
		return nil, nil, fmt.Errorf("consume %s: %w", QueueUnknownPerson, err)
	}
	return intruder, unknown, nil
}

func (c *Client) stopConsuming() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, tag := range []string{QueueIntruder, QueueUnknownPerson} {
		if err := c.channel.Cancel(tag, false); err != nil && !errors.Is(err, amqp.ErrClosed) {
			c.logger.Error("failed to cancel consumer", "consumer", tag, "error", err)
		}
	}
}

func (c *Client) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.connected.Store(false)

	var errs []error
	if err := c.channel.Close(); err != nil && !errors.Is(err, amqp.ErrClosed) {
		errs = append(errs, fmt.Errorf("failed to close channel: %w", err))
	}
	if err := c.conn.Close(); err != nil && !errors.Is(err, amqp.ErrClosed) {
		errs = append(errs, fmt.Errorf("failed to close connection: %w", err))
	}
	return errors.Join(errs...)
}
