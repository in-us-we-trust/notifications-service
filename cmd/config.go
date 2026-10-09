package main

import (
	"fmt"
	"notification-service/internal/fcm"
	"notification-service/internal/mailer"
	"notification-service/internal/rabbit"
	"notification-service/internal/users"
	"notification-service/lib/env"
	"notification-service/lib/toolbox"
)

type Config struct {
	Port         string         `json:"port"`
	RabbitConfig *rabbit.Config `json:"rabbitConfig"`
	FCMConfig    *fcm.Config    `json:"fcmConfig"`
	UsersConfig  *users.Config  `json:"usersConfig"`
	MailerConfig *mailer.Config `json:"mailerConfig"`
}

func loadConfig() (*Config, error) {
	err := env.LoadEnv()
	if err != nil {
		return nil, err
	}
	config := &Config{
		Port: env.GetString("PORT", "8080"),
		RabbitConfig: &rabbit.Config{
			URL:           toolbox.ToPointer(env.GetString("RABBIT_URL", "amqp://guest:guest@localhost:5672/")),
			PrefetchCount: toolbox.ToPointer(1),
		},
		FCMConfig: &fcm.Config{
			CredentialsPath: toolbox.ToPointer(env.GetString("FCM_CREDENTIALS_PATH", "")),
		},
		UsersConfig: &users.Config{
			BaseURL: toolbox.ToPointer(env.GetString("USERS_BASE_URL", "")),
		},
		MailerConfig: &mailer.Config{
			FromEmail:   env.GetString("MAILER_FROM_EMAIL", ""),
			FromPsw:     env.GetString("MAILER_FROM_PSW", ""),
			FrontendUrl: env.GetString("MAILER_FRONTEND_URL", ""),
			DevEmail:    env.GetString("MAILER_DEV_EMAIL", ""),
			Env:         env.GetString("ENV", ""),
		},
	}
	return config, nil
}

func (c *Config) validate() error {
	if c.RabbitConfig == nil {
		return fmt.Errorf("rabbitConfig is required")
	}
	if err := c.RabbitConfig.Validate(); err != nil {
		return fmt.Errorf("invalid rabbitConfig: %w", err)
	}
	if c.FCMConfig == nil {
		return fmt.Errorf("fcmConfig is required")
	}
	if err := c.FCMConfig.Validate(); err != nil {
		return fmt.Errorf("invalid fcmConfig: %w", err)
	}
	if c.UsersConfig == nil {
		return fmt.Errorf("usersConfig is required")
	}
	if err := c.UsersConfig.Validate(); err != nil {
		return fmt.Errorf("invalid usersConfig: %w", err)
	}
	if c.MailerConfig == nil {
		return fmt.Errorf("mailerConfig is required")
	}
	if err := c.MailerConfig.Validate(); err != nil {
		return fmt.Errorf("invalid mailerConfig: %w", err)
	}
	return nil
}
