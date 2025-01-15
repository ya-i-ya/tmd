package fetcher

import (
	"context"
	"fmt"

	"tmd/internal/db"
	"tmd/pkg/filehandler"

	"github.com/gotd/td/tg"
	"github.com/rs/zerolog/log"
)

type MeJob struct {
	MessageID      int
	TelegramUserID int64
	Media          tg.MessageMediaClass
	DialogName     string
}

func (f *Fetcher) workerMeJob() {
	defer f.wg.Done()
	for job := range f.meChan {
		if err := f.handleMeJob(job); err != nil {
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

	data, err := f.downloader.DownloadMediaToMemory(context.Background(), job.Media)
	if err != nil {
		return fmt.Errorf("download from Telegram to memory: %w", err)
	}

	objectName := filehandler.BuildObjectName(job.DialogName, mimeType, job.MessageID)

	mediaURL, err := f.storage.StoreBytes(context.Background(), data, mimeType, objectName)
	if err != nil {
		return fmt.Errorf("upload to MinIO: %w", err)
	}

	if err := f.database.Conn.Model(&db.Message{}).
		Where("message_id = ?", job.MessageID).
		Update("media_url", mediaURL).Error; err != nil {
		return fmt.Errorf("update database with media URL: %w", err)
	}

	log.Info().
		Int("message_id", job.MessageID).
		Str("media_url", mediaURL).
		Msg("Media uploaded to MinIO and database updated")
	return nil
}
