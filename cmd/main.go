package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"tmd/config"
	"tmd/internal"
)

func main() {
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to read config.yaml: %v", err)
	}

	client := telegram.NewClient(
		cfg.Telegram.ApiID,
		cfg.Telegram.ApiHash,
		telegram.Options{},
	)

	err = client.Run(context.Background(), func(ctx context.Context) error {
		if authErr := internal.EnsureAuth(ctx, client, cfg); authErr != nil {
			return authErr
		}
		fmt.Println("Client is now authorized!")
		tgClient := tg.NewClient(client)
		_ = tgClient

		return nil
	})
	if err != nil {
		log.Fatalf("Client run error: %v", err)
	}
}
