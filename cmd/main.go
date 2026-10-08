package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"notification-service/internal/fcm"
	"notification-service/internal/mailer"
	"notification-service/internal/rabbit"
	"notification-service/internal/users"
)

type Processor struct {
	usersClient  *users.Client
	fcmClient    *fcm.Client
	mailerClient *mailer.Mailer
	logger       *slog.Logger
}

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
	if err := config.validate(); err != nil {
		logger.Error("invalid config", "error", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	rabbitClient, err := rabbit.NewClient(*config.RabbitConfig, logger)
	if err != nil {
		logger.Error("failed to create RabbitMQ client", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := rabbitClient.Close(); err != nil {
			logger.Error("failed to close RabbitMQ client", "error", err)
		}
	}()

	processor, err := buildProcessor(ctx, config, logger)
	if err != nil {
		logger.Error("failed to build processor", "error", err)
		os.Exit(1)
	}

	healthDone := serveHealth(ctx, config.Port, rabbitClient, logger)

	run(ctx, rabbitClient, *config.RabbitConfig.Queue, processor, logger)

	<-healthDone
}

// serveHealth levanta el servidor HTTP de health check y lo apaga cuando se cancela ctx.
// El canal devuelto se cierra cuando el servidor terminó de apagarse.
func serveHealth(ctx context.Context, port string, rabbitClient *rabbit.Client, logger *slog.Logger) <-chan struct{} {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if !rabbitClient.IsConnected() {
			http.Error(w, "rabbit disconnected", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("health server listening", "port", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("health server failed", "error", err)
		}
	}()

	done := make(chan struct{})
	go func() {
		defer close(done)
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("failed to shutdown health server", "error", err)
		}
	}()

	return done
}

func buildProcessor(ctx context.Context, config *Config, logger *slog.Logger) (*Processor, error) {
	usersClient, err := users.NewClient(*config.UsersConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create users client: %w", err)
	}

	fcmClient, err := fcm.NewClient(ctx, *config.FCMConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create FCM client: %w", err)
	}

	mailerClient, err := mailer.NewMailer(*config.MailerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create mailer client: %w", err)
	}

	return &Processor{
		usersClient:  usersClient,
		fcmClient:    fcmClient,
		mailerClient: mailerClient,
		logger:       logger,
	}, nil
}

func run(ctx context.Context, rabbitClient *rabbit.Client, queue string, processor *Processor, logger *slog.Logger) {
	msgs := rabbitClient.Consume(ctx, queue)

	logger.Info("starting to consume messages")

	// El mensaje en curso se termina de procesar aunque llegue SIGTERM: Consume deja de
	// entregar mensajes nuevos y cierra msgs, y recién ahí sale el loop.
	processCtx := context.WithoutCancel(ctx)

	for msg := range msgs {
		if err := processor.process(processCtx, msg); err != nil {
			logger.Error("failed to process message", "error", err)
			msg.Ack()
			continue
		}
		msg.Ack()
	}

	logger.Info("shutdown complete")
}

func (p *Processor) process(ctx context.Context, msg rabbit.Message) error {
	var notification rabbit.Notification
	if err := json.Unmarshal(msg.Body, &notification); err != nil {
		return fmt.Errorf("failed to unmarshal notification: %w", err)
	}

	p.logger.Info("processing notification",
		"deviceId", notification.DeviceID,
		"title", notification.Title,
	)

	contacts, err := p.usersClient.GetUsersByDevice(ctx, notification.DeviceID)
	if err != nil {
		return fmt.Errorf("failed to get contacts for device %s: %w", notification.DeviceID, err)
	}

	p.logger.Info("sending notifications", "deviceId", notification.DeviceID, "contacts", len(contacts))

	for _, user := range contacts {
		if len(user.NotificationTokens) == 0 {
			if err := p.mailerClient.SendNotification(user.Email, user.Name, notification.Title, notification.Message, notification.ImageURL); err != nil {
				p.logger.Error("failed to send email",
					"userId", user.UserID,
					"email", user.Email,
					"error", err,
				)
			} else {
				p.logger.Info("email sent", "userId", user.UserID, "email", user.Email)
			}
			continue
		}

		for _, nt := range user.NotificationTokens {
			if err := p.fcmClient.Send(ctx, nt.Token, notification.Title, notification.Message); err != nil {
				p.logger.Error("failed to send push notification",
					"userId", user.UserID,
					"deviceName", nt.DeviceName,
					"error", err,
				)

				// TODO: Borro siempre a modo de prueba
				// if errors.Is(err, fcm.ErrTokenExpired) {
				p.logger.Warn("FCM token expired, removing",
					"userId", user.UserID,
					"appDeviceId", nt.AppDeviceID,
				)
				if err := p.usersClient.DeletePushToken(ctx, user.UserID, nt.AppDeviceID); err != nil {
					p.logger.Error("failed to delete expired push token",
						"userId", user.UserID,
						"appDeviceId", nt.AppDeviceID,
						"error", err,
					)
				}
				// }

				if err := p.mailerClient.SendNotification(user.Email, user.Name, notification.Title, notification.Message, notification.ImageURL); err != nil {
					p.logger.Error("failed to send fallback email",
						"userId", user.UserID,
						"email", user.Email,
						"error", err,
					)
				} else {
					p.logger.Info("fallback email sent", "userId", user.UserID, "email", user.Email)
				}

				continue
			}
			p.logger.Info("notification sent", "userId", user.UserID, "deviceName", nt.DeviceName)
		}
	}

	return nil
}
