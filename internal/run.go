package internal

import (
	"context"
	"time"

	"tmd/internal/db"
	"tmd/internal/fetcher"
	"tmd/minio"
	"tmd/pkg/cfg"
	"tmd/pkg/filehandler"
	"tmd/pkg/logger"

	"github.com/gotd/td/telegram"
	"github.com/rs/zerolog/log"
)

func Run() error {
	if err := logger.SetupLogger("tmd.log", "info"); err != nil {
		log.Fatal().Err(err).Msg("Failed to setup logger")
	}

	config, err := cfg.LoadConfig("config.yaml")
	if err != nil {
		log.Error().Err(err).Msg("Failed to read config.yaml")
		return err
	}

	dbConn, err := db.NewDB(config)
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to DB")
		return err
	}

	st, err := minio.NewMinIOStorage(
		config.Minio.Endpoint,
		config.Minio.AccessKey,
		config.Minio.SecretKey,
		config.Minio.Bucket,
		config.Minio.BasePath,
		config.Minio.UseSSL,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create MinIO storage")
		return err
	}

	client := telegram.NewClient(
		config.Telegram.ApiID,
		config.Telegram.ApiHash,
		telegram.Options{},
	)

	downloader := filehandler.NewDownloader(client, config.Download.BaseDir)
	f := fetcher.NewFetcher(
		client,
		downloader,
		dbConn,
		st,
		config.Fetching.DialogsLimit,
		config.Fetching.MessagesLimit,
	)

	return client.Run(context.Background(), func(ctx context.Context) error {
		if err := EnsureAuth(ctx, client, config); err != nil {
			return err
		}
		log.Info().Msg("Client is authorized and ready!")

		go func() {
			defer f.CloseWorkers()
			for {
				if err := f.FetchAllDMs(ctx); err != nil {
					log.Error().Err(err).Msg("Failed to fetch DMs")
				}
				time.Sleep(5 * time.Minute)
			}
		}()

		select {}
	})
}
