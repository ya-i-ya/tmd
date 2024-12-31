package main

import (
	"context"
	"github.com/gotd/td/telegram"
	"github.com/sirupsen/logrus"
	"log"
	"tmd/config"
	"tmd/internal"
	"tmd/internal/logger"
)

func main() {
	if err := run(); err != nil {
		logrus.WithError(err).Fatal("Application failed")
	}
}
func run() error {

	if err := logger.SetupLogger("app.log"); err != nil {
		log.Fatalf("Failed to setup logger: %v", err)
	}

	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		logrus.WithError(err).Error("Failed to read config.yaml")
		return err
	}

	client := telegram.NewClient(
		cfg.Telegram.ApiID,
		cfg.Telegram.ApiHash,
		telegram.Options{},
	)

	return client.Run(context.Background(), func(ctx context.Context) error {
		if err := internal.EnsureAuth(ctx, client, cfg); err != nil {
			return err
		}
		logrus.Println("Client is authorized and ready!")
		return nil
	})
}
