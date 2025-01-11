package fetcher

import (
	"context"
	"fmt"
	"tmd/internal/db"

	"github.com/gotd/td/tg"
	"github.com/rs/zerolog/log"
	"tmd/pkg/filehandler"
)

type MeJob struct {
	MessageID      int
	TelegramUserID int64
	ChatID         int64
	Media          tg.MessageMediaClass
	MediaFilePath  string
	MediaURL       string
}

func (f *Fetcher) workerMeJob() {
	defer f.wg.Done()
	for job := range f.meChan {
		err := f.handleMeJob(job)
		if err != nil {
			log.Error().
				Err(err).
				Int("message_id", job.MessageID).
				Msg("Failed to handle media job")
		}
	}
}

func (f *Fetcher) handleMeJob(job MeJob) error {
	mimeType, err := filehandler.GetMimeType(job.Media)
	if err != nil {
		return fmt.Errorf("determine MIME type: %w", err)
	}

	mediaURL, err := f.storage.StoreFile(context.Background(), job.MediaFilePath, mimeType)
	if err != nil {
		return fmt.Errorf("upload to MinIO: %w", err)
	}
	job.MediaURL = mediaURL

	err = f.database.Conn.Model(&db.Message{}).
		Where("message_id = ?", job.MessageID).
		Update("media_url", mediaURL).Error

	if err != nil {
		return fmt.Errorf("update database with media URL: %w", err)
	}

	log.Info().
		Int("message_id", job.MessageID).
		Str("media_url", mediaURL).
		Msg("Media uploaded to MinIO and database updated")

	return nil
}
