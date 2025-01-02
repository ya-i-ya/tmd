package internal

import (
	"context"
	"github.com/rs/zerolog/log"
	"tmd/logger"

	"github.com/gotd/td/telegram"
	"tmd/cfg"
)

func Run() error {
	if err := logger.SetupLogger("tmd.log"); err != nil {
		log.Fatal().Err(err).Msgf("Failed to setup logger: %v", err)
	}

	cfg, err := cfg.LoadConfig("config.yaml")
	if err != nil {
		log.Error().Err(err).Msg("Failed to read config.yaml")
		return err
	}

	client := telegram.NewClient(
		cfg.Telegram.ApiID,
		cfg.Telegram.ApiHash,
		telegram.Options{},
	)

	return client.Run(context.Background(), func(ctx context.Context) error {
		if err := EnsureAuth(ctx, client, cfg); err != nil {
			return err
		}
		log.Info().Msg("Client is authorized and ready!")
		return nil
	})
}
