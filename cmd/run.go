package main

//
//import (
//	"context"
//	"github.com/sirupsen/logrus"
//	"log"
//
//	"github.com/gotd/td/telegram"
//	"tmd/cfg"
//	"tmd/internal"
//	"tmd/internal/logger"
//)
//
//func run() error {
//
//	if err := logger.SetupLogger("app.log"); err != nil {
//		log.Fatalf("Failed to setup logger: %v", err)
//	}
//
//	cfg, err := cfg.LoadConfig("configs/cfg.yaml")
//	if err != nil {
//		logrus.WithError(err).Error("Failed to read cfg.yaml")
//	}
//
//	client := telegram.NewClient(
//		cfg.Telegram.ApiID,
//		cfg.Telegram.ApiHash,
//		telegram.Options{},
//	)
//
//	return client.Run(context.Background(), func(ctx context.Context) error {
//		if err := internal.EnsureAuth(ctx, client, cfg); err != nil {
//			return err
//		}
//		logrus.Println("Client is authorized and ready!")
//		return nil
//	})
//}
