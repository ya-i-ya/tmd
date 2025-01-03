package filehandler

import (
	"context"
	"fmt"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/rs/zerolog/log"
)

type Downloader struct {
	client  *telegram.Client
	baseDir string
}

func NewDownloader(client *telegram.Client, baseDir string) *Downloader {
	return &Downloader{
		client:  client,
		baseDir: baseDir,
	}
}

func (d *Downloader) ProcessMedia(ctx context.Context, messageID int, media tg.MessageMediaClass, chatID int) error {
	switch m := media.(type) {
	case *tg.MessageMediaPhoto:
		return d.DownloadPhoto(ctx, messageID, m, chatID)
	case *tg.MessageMediaDocument:
		return d.DownloadDocument(ctx, messageID, m, chatID)
	default:
		log.Warn().
			Str("media_type", fmt.Sprintf("%T", m)).
			Int("message_id", messageID).
			Msg("Unsupported media type")
		return nil
	}
}

func (d *Downloader) DownloadPhoto(ctx context.Context, messageID int, media *tg.MessageMediaPhoto, chatID int) error {
	//
	log.Info().
		Int("message_id", messageID).
		Str("media_type", "photo").
		Msg("Downloading photo")
	return nil
}

func (d *Downloader) DownloadDocument(ctx context.Context, messageID int, media *tg.MessageMediaDocument, chatID int) error {
	//
	log.Info().
		Int("message_id", messageID).
		Str("media_type", "document").
		Msg("Downloading document")
	return nil
}
