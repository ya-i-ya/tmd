package internal

import (
	"context"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/rs/zerolog/log"
	"tmd/cfg"
	"tmd/logger"
)

func Run() error {
	if err := logger.SetupLogger(cfgPath(), "info"); err != nil {
		log.Fatal().Err(err).Msgf("Failed to setup logger: %v", err)
	}

	config, err := cfg.LoadConfig("config.yaml")
	if err != nil {
		log.Error().Err(err).Msg("Failed to read config.yaml")
		return err
	}

	client := telegram.NewClient(
		config.Telegram.ApiID,
		config.Telegram.ApiHash,
		telegram.Options{},
	)

	downloader := NewDownloader(client, config.Download.BaseDir)
	fetcher := NewFetcher(client, downloader, config.Fetching.DialogsLimit, config.Fetching.MessagesLimit)

	return client.Run(context.Background(), func(ctx context.Context) error {
		if err := EnsureAuth(ctx, client, config); err != nil {
			return err
		}
		log.Info().Msg("Client is authorized and ready!")

		go func() {
			for {
				if err := fetcher.FetchAllDMs(ctx); err != nil {
					log.Error().
						Err(err).
						Msg("Failed to fetch DMs")
				}
				time.Sleep(5 * time.Minute)
			}
		}()

		select {}
	})
}

func cfgPath() string {
	return "config.yaml"
}
