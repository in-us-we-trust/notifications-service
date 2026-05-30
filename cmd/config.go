package main

import (
	"notification-service/internal/rabbit"
	"notification-service/lib/env"
	"notification-service/lib/toolbox"
)

type Config struct {
	RabbitConfig *rabbit.Config `json:"rabbitConfig"`
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
	}
	return config, nil
}
