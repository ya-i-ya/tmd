package internal

import (
	"context"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/sirupsen/logrus"
	"os"
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
func (d *Downloader) ProcessMedia(ctx context.Context, messageID int, media tg.MessageMediaClass) error {
	switch m := media.(type) {
	case *tg.MessageMediaPhoto:
		return d.DownloadPhoto(ctx, messageID, m)
	case *tg.MessageMediaDocument:
		return d.DownloadDocument(ctx, messageID, m)
	default:
		logrus.Warnf("Unsupported media type: %T for message ID: %d", m, messageID)
		return nil
	}
}
func (d *Downloader) DownloadPhoto(ctx context.Context, messageID int, media *tg.MessageMediaPhoto) error {
	//
	return nil
}

func (d *Downloader) DownloadDocument(ctx context.Context, messageID int, media *tg.MessageMediaDocument) error {
	//
	return nil
}

func ensureDir(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}
