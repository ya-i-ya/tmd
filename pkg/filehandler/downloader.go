package filehandler

import (
	"bytes"
	"context"
	"fmt"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/downloader"
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

func (d *Downloader) DownloadMediaToMemory(ctx context.Context, media tg.MessageMediaClass) ([]byte, error) {
	switch m := media.(type) {
	case *tg.MessageMediaPhoto:
		return d.downloadPhotoToMemory(ctx, m)
	case *tg.MessageMediaDocument:
		return d.downloadDocumentToMemory(ctx, m)
	default:
		log.Warn().Str("media_type", fmt.Sprintf("%T", m)).Msg("Unsupported media type")
		return nil, nil
	}
}

func (d *Downloader) downloadPhotoToMemory(ctx context.Context, media *tg.MessageMediaPhoto) ([]byte, error) {
	photo := media.Photo
	if photo == nil {
		log.Warn().Msg("Skipping empty photo")
		return nil, nil
	}
	photoObj, ok := photo.(*tg.Photo)
	if !ok || photoObj == nil {
		log.Warn().Msg("Photo object is invalid")
		return nil, nil
	}

	var chosenThumb *tg.PhotoSize
	for i := len(photoObj.Sizes) - 1; i >= 0; i-- {
		if sz, ok := photoObj.Sizes[i].(*tg.PhotoSize); ok {
			chosenThumb = sz
			break
		}
	}

	if chosenThumb == nil {
		log.Warn().Msg("No suitable photo size found")
		return nil, nil
	}

	location := &tg.InputPhotoFileLocation{
		ID:            photoObj.ID,
		AccessHash:    photoObj.AccessHash,
		FileReference: photoObj.FileReference,
		ThumbSize:     chosenThumb.Type,
	}

	dl := downloader.NewDownloader().WithPartSize(1024 * 1024)

	var buffer bytes.Buffer
	if _, err := dl.Download(d.client.API(), location).Stream(ctx, &buffer); err != nil {
		return nil, fmt.Errorf("failed to download photo: %w", err)
	}
	return buffer.Bytes(), nil
}

func (d *Downloader) downloadDocumentToMemory(ctx context.Context, media *tg.MessageMediaDocument) ([]byte, error) {
	doc := media.Document
	if doc == nil {
		log.Warn().Msg("Skipping empty document")
		return nil, nil
	}

	docObj, ok := doc.(*tg.Document)
	if !ok || docObj == nil {
		log.Warn().Msg("Document object is invalid")
		return nil, nil
	}

	location := &tg.InputDocumentFileLocation{
		ID:            docObj.ID,
		AccessHash:    docObj.AccessHash,
		FileReference: docObj.FileReference,
		ThumbSize:     "",
	}

	dl := downloader.NewDownloader().WithPartSize(1024 * 1024)

	var buffer bytes.Buffer
	if _, err := dl.Download(d.client.API(), location).Stream(ctx, &buffer); err != nil {
		return nil, fmt.Errorf("failed to download document: %w", err)
	}
	return buffer.Bytes(), nil
}
