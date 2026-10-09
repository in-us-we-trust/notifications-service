package fcm

import (
	"context"
	"fmt"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type Client struct {
	messaging *messaging.Client
}

func NewClient(ctx context.Context, config Config) (*Client, error) {
	credBytes, err := os.ReadFile(*config.CredentialsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	app, err := firebase.NewApp(ctx, &firebase.Config{
		ProjectID: "tp-hogar-seguro",
	}, option.WithCredentialsJSON(credBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create firebase app: %w", err)
	}

	msgClient, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create messaging client: %w", err)
	}

	return &Client{messaging: msgClient}, nil
}

func (c *Client) Send(ctx context.Context, token string, title string, body string, imageURL string, data map[string]string) error {
	message := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title:    title,
			Body:     body,
			ImageURL: imageURL,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
	}
	if imageURL != "" {
		message.Android.Notification = &messaging.AndroidNotification{
			ImageURL: imageURL,
		}
		message.APNS = &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					MutableContent: true,
				},
			},
			FCMOptions: &messaging.APNSFCMOptions{
				ImageURL: imageURL,
			},
		}
	}

	_, err := c.messaging.Send(ctx, message)
	if err != nil {
		if messaging.IsUnregistered(err) {
			return ErrTokenExpired
		}
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}
