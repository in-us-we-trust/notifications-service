package rabbit

import (
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Config struct {
	URL           *string `json:"url"`
	PrefetchCount *int    `json:"prefetchCount"`
	Queue         *string `json:"queue"`
}

func (c *Config) validate() error {
	if c.URL == nil || *c.URL == "" {
		return fmt.Errorf("url is required")
	}
	if c.PrefetchCount == nil {
		return fmt.Errorf("prefetchCount is required")
	}
	if c.Queue == nil || *c.Queue == "" {
		return fmt.Errorf("queue is required")
	}
	return nil
}

type Message struct {
	Body     []byte
	delivery amqp.Delivery
}

func (m *Message) Ack() error {
	return m.delivery.Ack(false)
}

func (m *Message) Nack(requeue bool) error {
	return m.delivery.Nack(false, requeue)
}

type Notification struct {
	Severity string    `json:"severity"`
	Tags     []string  `json:"tags"`
	DeviceID string    `json:"deviceId"`
	Time     time.Time `json:"time"`
	Title    string    `json:"title"`
	Message  string    `json:"message"`
}
