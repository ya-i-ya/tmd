package main

import (
	"context"
	"github.com/gotd/td/telegram"
	"tmd/cfg"
	"tmd/internal"
	"tmd/internal/logger"

	"github.com/rs/zerolog/log"
)

func main() {
	if err := run(); err != nil {
		log.Fatal().Err(err).Msg("Application failed")
	}
}

func run() error {
	if err := logger.SetupLogger("app.log"); err != nil {
		log.Fatal().Err(err).Msgf("Failed to setup logger: %v", err)
	}

	cfg, err := cfg.LoadConfig("example.config.yaml")
	if err != nil {
		log.Error().Err(err).Msg("Failed to read cfg.yaml")
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
		log.Info().Msg("Client is authorized and ready!")
		return nil
	})
}
