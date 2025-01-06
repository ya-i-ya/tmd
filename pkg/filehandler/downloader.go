package filehandler

import (
	"context"
	"fmt"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
	"github.com/rs/zerolog/log"
)

type Downloader struct {
	client    *telegram.Client
	baseDir   string
	organizer *Organizer
}

func NewDownloader(client *telegram.Client, baseDir string) *Downloader {
	return &Downloader{
		client:    client,
		baseDir:   baseDir,
		organizer: NewOrganizer(baseDir),
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
	log.Info().
		Int("message_id", messageID).
		Str("media_type", "photo").
		Msg("Downloading photo")

	photo := media.Photo
	if photo == nil {
		log.Warn().Int("message_id", messageID).Msg("Skipping empty photo")
		return nil
	}
	photoObj, ok := photo.(*tg.Photo)
	if !ok {
		log.Warn().Int("message_id", messageID).Msg("Photo object is not a valid *tg.Photo")
		return nil
	}
	sizes := photoObj.Sizes
	if len(sizes) == 0 {
		log.Warn().Int("message_id", messageID).Msg("No photo sizes found")
		return nil
	}
	var chosen *tg.PhotoSize
	for i := len(sizes) - 1; i >= 0; i-- {
		if sz, ok := sizes[i].(*tg.PhotoSize); ok {
			chosen = sz
			break
		}
	}
	if chosen == nil {
		log.Warn().
			Int("message_id", messageID).
			Msg("Last photo size is not a *tg.PhotoSize, skipping download")
		return nil
	}
	location := &tg.InputPhotoFileLocation{
		ID:            photoObj.ID,
		AccessHash:    photoObj.AccessHash,
		FileReference: photoObj.FileReference,
		ThumbSize:     chosen.Type,
	}
	dl := downloader.NewDownloader().WithPartSize(512 * 1024)
	mimeType := "image/jpeg"
	path, err := d.organizer.getFilePath(mimeType, chatID, messageID)
	if err != nil {
		return fmt.Errorf("failed to get path: %w", err)
	}
	_, err = dl.Download(d.client.API(), location).ToPath(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to download photo: %w", err)
	}
	log.Info().
		Str("file_path", path).
		Int("message_id", messageID).
		Msg("Downloaded and saved photo successfully")
	return nil
}

func (d *Downloader) DownloadDocument(ctx context.Context, messageID int, media *tg.MessageMediaDocument, chatID int) error {
	log.Info().
		Int("message_id", messageID).
		Str("media_type", "document").
		Msg("Downloading document")

	doc := media.Document
	if doc == nil {
		log.Warn().Int("message_id", messageID).Msg("Skipping empty document")
		return nil
	}

	docObj, ok := doc.(*tg.Document)
	if !ok {
		log.Warn().Int("message_id", messageID).Msg("Document object is not a valid *tg.Document")
		return nil
	}

	location := &tg.InputDocumentFileLocation{
		ID:            docObj.ID,
		AccessHash:    docObj.AccessHash,
		FileReference: docObj.FileReference,
		ThumbSize:     "",
	}

	dl := downloader.NewDownloader().WithPartSize(512 * 1024)

	mimeType := docObj.MimeType
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	path, err := d.organizer.getFilePath(mimeType, chatID, messageID)
	if err != nil {
		return fmt.Errorf("failed to get path: %w", err)
	}
	_, err = dl.Download(d.client.API(), location).ToPath(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to download document: %w", err)
	}
	log.Info().
		Str("file_path", path).
		Int("message_id", messageID).
		Msg("Downloaded and saved document successfully")
	return nil
}
