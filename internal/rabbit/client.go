package rabbit

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  Config
	mutex   sync.Mutex
	logger  *slog.Logger
}

func NewClient(config Config, logger *slog.Logger) (*Client, error) {
	c := &Client{config: config, logger: logger}
	if err := c.connect(); err != nil {
		return nil, err
	}

	return c, nil
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

	c.conn = conn
	c.channel = ch
	return nil
}

func (c *Client) reconnect(ctx context.Context) error {
	backoff := 2 * time.Second
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("reconnect cancelled")
		default:
		}

		c.logger.Info("reconnecting...", "backoff", backoff)
		time.Sleep(backoff)

		if err := c.connect(); err != nil {
			c.logger.Error("reconnect failed", "error", err)
			backoff = min(backoff*2, 60*time.Second)
			continue
		}

		c.logger.Info("reconnected successfully")
		return nil
	}
}

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

			for msg := range msgs {
				select {
				case out <- Message{Body: msg.Body, delivery: msg}:
				case <-ctx.Done():
					return
				}
			}

			c.logger.Warn("connection lost, reconnecting...")
			if err := c.reconnect(ctx); err != nil {
				return
			}
		}
	}()

	return out
}

func (c *Client) startConsuming(queue string) (<-chan amqp.Delivery, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.channel.Consume(queue, "", false, false, false, false, nil)
}

func (c *Client) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if err := c.channel.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %w", err)
	}
	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}
	return nil
}
