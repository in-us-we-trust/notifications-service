package main

import (
	"context"
	"os"
	"os/signal"

	"notification-service/internal/rabbit"

	"log/slog"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	config, err := loadConfig()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	rabbitClient, err := rabbit.NewClient(*config.RabbitConfig, logger)
	if err != nil {
		logger.Error("failed to create RabbitMQ client", "error", err)
		os.Exit(1)
	}
	defer rabbitClient.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	msgs := rabbitClient.Consume(ctx, *config.RabbitConfig.Queue)

	logger.Info("starting to consume messages")

	for msg := range msgs {
		if err := process(msg); err != nil {
			logger.Error("failed to process message", "error", err)
			msg.Nack(true)
			continue
		}
		msg.Ack()
	}

	logger.Info("shutdown complete")
}

func process(msg rabbit.Message) error {
	slog.Info("received message", "body", string(msg.Body))
	return nil
}
