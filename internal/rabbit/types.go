package rabbit

import (
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ExchangeNotifications = "notifications"
	QueueIntruder         = "notification.intruder"
	QueueUnknownPerson    = "notification.unknown-person"
)

type Config struct {
	URL           *string `json:"url"`
	PrefetchCount *int    `json:"prefetchCount"`
}

func (c *Config) Validate() error {
	if c.URL == nil || *c.URL == "" {
		return fmt.Errorf("url is required")
	}
	if c.PrefetchCount == nil {
		return fmt.Errorf("prefetchCount is required")
	}
	return nil
}

type Message struct {
	Queue    string
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
	ImageURL string    `json:"imageUrl,omitempty"`
}

type UnknownPersonSuggestion struct {
	DeviceID     string    `json:"deviceId"`
	PersonID     string    `json:"personId"`
	Time         time.Time `json:"time"`
	DistinctDays int       `json:"distinctDays"`
	ImageURLs    []string  `json:"imageUrls"`
}

type PushContent struct {
	Type     string
	DeviceID string
	PersonID string
	Title    string
	Body     string
	ImageURL string
	Data     map[string]string
}
