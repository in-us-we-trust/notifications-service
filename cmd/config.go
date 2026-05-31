package main

import (
	"notification-service/internal/fcm"
	"notification-service/internal/rabbit"
	"notification-service/internal/users"
	"notification-service/lib/env"
	"notification-service/lib/toolbox"
)

type Config struct {
	RabbitConfig *rabbit.Config `json:"rabbitConfig"`
	FCMConfig    *fcm.Config    `json:"fcmConfig"`
	UsersConfig  *users.Config  `json:"usersConfig"`
}

func loadConfig() (*Config, error) {
	err := env.LoadEnv()
	if err != nil {
		return nil, err
	}
	config := &Config{
		RabbitConfig: &rabbit.Config{
			URL:           toolbox.ToPointer(env.GetString("RABBIT_URL", "amqp://guest:guest@localhost:5672/")),
			PrefetchCount: toolbox.ToPointer(1),
			Queue:         toolbox.ToPointer(env.GetString("RABBIT_QUEUE", "")),
		},
		FCMConfig: &fcm.Config{
			CredentialsPath: toolbox.ToPointer(env.GetString("FCM_CREDENTIALS_PATH", "")),
		},
		UsersConfig: &users.Config{
			BaseURL: toolbox.ToPointer(env.GetString("USERS_BASE_URL", "")),
		},
	}
	return config, nil
}
