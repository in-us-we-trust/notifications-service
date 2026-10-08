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

const consumerTag = "notification-service"

type Client struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	config    Config
	mutex     sync.Mutex
	connected atomic.Bool
	logger    *slog.Logger
}

func NewClient(config Config, logger *slog.Logger) (*Client, error) {
	c := &Client{config: config, logger: logger}
	if err := c.connect(); err != nil {
		return nil, err
	}

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

	c.mutex.Lock()
	c.conn = conn
	c.channel = ch
	c.mutex.Unlock()

	// Se marca como desconectado apenas se cae la conexión o el canal, sin esperar
	// a que el consumer lo detecte.
	connClosed := conn.NotifyClose(make(chan *amqp.Error, 1))
	chClosed := ch.NotifyClose(make(chan *amqp.Error, 1))
	go func() {
		select {
		case <-connClosed:
		case <-chClosed:
		}
		c.connected.Store(false)
	}()

	c.connected.Store(true)
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

// Consume entrega los mensajes de la cola hasta que se cancela el context. Al cancelarse,
// deja de consumir (cancela el consumer en el broker) y cierra el canal devuelto.
func (c *Client) Consume(ctx context.Context, queue string) <-chan Message {
	out := make(chan Message)

	go func() {
		defer close(out)
		for {
			msgs, err := c.startConsuming(queue)
			if err != nil {
				c.logger.Error("failed to start consuming", "error", err)
				if err := c.reconnect(ctx); err != nil {
					return
				}
				continue
			}

			if stopped := c.forward(ctx, msgs, out); stopped {
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
// context y false si el canal de deliveries se cerró (conexión caída).
func (c *Client) forward(ctx context.Context, msgs <-chan amqp.Delivery, out chan<- Message) bool {
	for {
		select {
		case <-ctx.Done():
			c.stopConsuming()
			return true
		case msg, ok := <-msgs:
			if !ok {
				return false
			}
			select {
			case out <- Message{Body: msg.Body, delivery: msg}:
			case <-ctx.Done():
				// No llegó a procesarse: se devuelve a la cola.
				if err := msg.Nack(false, true); err != nil {
					c.logger.Error("failed to requeue message on shutdown", "error", err)
				}
				c.stopConsuming()
				return true
			}
		}
	}
}

func (c *Client) startConsuming(queue string) (<-chan amqp.Delivery, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.channel.Consume(queue, consumerTag, false, false, false, false, nil)
}

func (c *Client) stopConsuming() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if err := c.channel.Cancel(consumerTag, false); err != nil && !errors.Is(err, amqp.ErrClosed) {
		c.logger.Error("failed to cancel consumer", "error", err)
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
