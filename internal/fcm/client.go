package fcm

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type Client struct {
	messaging *messaging.Client
}

func NewClient(ctx context.Context, config Config) (*Client, error) {
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsFile(*config.CredentialsPath))
	if err != nil {
		return nil, fmt.Errorf("failed to create firebase app: %w", err)
	}

	msgClient, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create messaging client: %w", err)
	}

	return &Client{messaging: msgClient}, nil
}

func (c *Client) Send(ctx context.Context, token string, title string, body string) error {
	message := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
	}

	_, err := c.messaging.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}
