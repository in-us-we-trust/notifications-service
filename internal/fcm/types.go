package fcm

import (
	"errors"
	"fmt"
)

var ErrTokenExpired = errors.New("FCM token expired or unregistered")

type Config struct {
	CredentialsPath *string `json:"credentialsPath"`
}

func (c *Config) Validate() error {
	if c.CredentialsPath == nil || *c.CredentialsPath == "" {
		return fmt.Errorf("credentials path is required")
	}
	return nil
}
