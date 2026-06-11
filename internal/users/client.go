// internal/users/client.go
package users

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	httpClient *http.Client
	config     Config
}

func NewClient(config Config) (*Client, error) {
	return &Client{
		httpClient: &http.Client{},
		config:     config,
	}, nil
}

func (c *Client) GetUsersByDevice(ctx context.Context, deviceID string) ([]User, error) {
	url := fmt.Sprintf("%s/devices/%s/contacts", *c.config.BaseURL, deviceID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var users []User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return users, nil
}

func (c *Client) DeletePushToken(ctx context.Context, userID string, appDeviceID int) error {
	url := fmt.Sprintf("%s/users/%s/app-devices/%d/push-token", *c.config.BaseURL, userID, appDeviceID)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
