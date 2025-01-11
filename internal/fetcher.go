package internal

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"sync"
	"tmd/internal/db"
	"tmd/minio"
	"tmd/pkg/filehandler"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/rs/zerolog/log"
)

type Fetcher struct {
	client        *telegram.Client
	downloader    *filehandler.Downloader
	database      *db.DB
	storage       *minio.Storage
	dialogsLimit  int
	messagesLimit int

	meChan chan MeJob // ^.^
	wg     sync.WaitGroup
}

type MeJob struct {
	MessageID     int
	Media         tg.MessageMediaClass
	ChatID        int
	UserID        uuid.UUID
	MediaFilePath string
	MediaURL      string
}

func NewFetcher(client *telegram.Client, downloader *filehandler.Downloader, database *db.DB, storage *minio.Storage, dialogsLimit, messagesLimit int) *Fetcher {
	f := &Fetcher{
		client:        client,
		downloader:    downloader,
		database:      database,
		storage:       storage,
		dialogsLimit:  dialogsLimit,
		messagesLimit: messagesLimit,
		meChan:        make(chan MeJob, 100),
	}

	workerCount := 5
	for i := 0; i < workerCount; i++ {
		f.wg.Add(1)
		go f.meJob()
	}

	return f
}
func (f *Fetcher) meJob() {
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
		Msg("Media uploaded to MinIO and database updated successfully")

	return nil
}

func (f *Fetcher) FetchAllDMs(ctx context.Context) error {
	tgClient := tg.NewClient(f.client)
	offsetDate, offsetID := 0, 0
	var offsetPeer tg.InputPeerClass = &tg.InputPeerEmpty{}

	for {
		res, err := tgClient.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
			OffsetDate: offsetDate,
			OffsetID:   offsetID,
			OffsetPeer: offsetPeer,
			Limit:      f.dialogsLimit,
			Hash:       0,
		})
		if err != nil {
			return fmt.Errorf("failed to get dialogs: %w", err)
		}

		switch d := res.(type) {
		case *tg.MessagesDialogsSlice:
			for _, dialog := range d.Dialogs {
				if err := f.processDialog(ctx, dialog, d.Users); err != nil {
					log.Warn().
						Err(err).
						Msg("Failed to process dialog")
				}
			}
			if len(d.Dialogs) < f.dialogsLimit {
				return nil
			}
			last := d.Dialogs[len(d.Dialogs)-1]
			offsetPeer = f.getNextOffsetPeer(last)
			offsetID = 0
			offsetDate = 0

		case *tg.MessagesDialogs:
			for _, dialog := range d.Dialogs {
				if err := f.processDialog(ctx, dialog, d.Users); err != nil {
					log.Warn().
						Err(err).
						Msg("Failed to process dialog")
				}
			}
			return nil

		default:
			log.Warn().
				Str("type", fmt.Sprintf("%T", res)).
				Msg("Unexpected dialog type")
			return nil
		}
	}
}

func (f *Fetcher) processDialog(ctx context.Context, dialog tg.DialogClass, users []tg.UserClass) error {
	d, ok := dialog.(*tg.Dialog)
	if !ok {
		log.Warn().
			Str("dialog_type", fmt.Sprintf("%T", dialog)).
			Msg("Skipping unsupported dialog type")
		return nil
	}

	switch peer := d.Peer.(type) {
	case *tg.PeerUser:
		user, found := findUser(users, peer.UserID)
		if !found {
			log.Warn().
				Int64("user_id", peer.UserID).
				Msg("User not found for peer ID")
			return nil
		}
		inputPeer := &tg.InputPeerUser{
			UserID:     user.ID,
			AccessHash: user.AccessHash,
		}
		chatID := int(user.ID)
		return f.FetchAndProcessMessages(ctx, inputPeer, chatID)
	default:
		log.Warn().
			Str("peer_type", fmt.Sprintf("%T", peer)).
			Msg("Unsupported peer type")
		return nil
	}
}

func (f *Fetcher) getNextOffsetPeer(dialog tg.DialogClass) tg.InputPeerClass {
	d, ok := dialog.(*tg.Dialog)
	if !ok {
		log.Warn().
			Str("dialog_type", fmt.Sprintf("%T", dialog)).
			Msg("Unexpected dialog type when getting next offset peer")
		return &tg.InputPeerEmpty{}
	}

	peer := d.GetPeer()
	if peer == nil {
		log.Warn().
			Msg("Peer is nil, setting offsetPeer to InputPeerEmpty")
		return &tg.InputPeerEmpty{}
	}

	return dialogToInputPeer(peer)
}

func findUser(users []tg.UserClass, userID int64) (*tg.User, bool) {
	for _, u := range users {
		if user, ok := u.(*tg.User); ok && user.ID == userID {
			return user, true
		}
	}
	return nil, false
}

func dialogToInputPeer(peer tg.PeerClass) tg.InputPeerClass {
	switch p := peer.(type) {
	case *tg.PeerUser:
		return &tg.InputPeerUser{UserID: p.UserID}
	case *tg.PeerChat:
		return &tg.InputPeerChat{ChatID: p.ChatID}
	case *tg.PeerChannel:
		return &tg.InputPeerChannel{ChannelID: p.ChannelID}
	default:
		return &tg.InputPeerEmpty{}
	}
}

func (f *Fetcher) FetchAndProcessMessages(ctx context.Context, peer tg.InputPeerClass, chatID int) error {
	tgClient := tg.NewClient(f.client)
	offsetID := 0

	for {
		history, err := tgClient.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
			Peer:     peer,
			OffsetID: offsetID,
			Limit:    f.messagesLimit,
		})

		if err != nil {
			return fmt.Errorf("failed to f message history: %w", err)
		}

		switch msgs := history.(type) {
		case *tg.MessagesChannelMessages:
			if len(msgs.Messages) == 0 {
				return nil
			}
			if err := f.processMessagesBatch(ctx, msgs.Messages, &offsetID, chatID); err != nil {
				return err
			}
			if len(msgs.Messages) < f.messagesLimit {
				return nil
			}

		case *tg.MessagesMessagesSlice:
			if len(msgs.Messages) == 0 {
				return nil
			}
			if err := f.processMessagesBatch(ctx, msgs.Messages, &offsetID, chatID); err != nil {
				return err
			}
			if len(msgs.Messages) < f.messagesLimit {
				return nil
			}

		case *tg.MessagesMessages:
			if len(msgs.Messages) == 0 {
				return nil
			}
			if err := f.processMessagesBatch(ctx, msgs.Messages, &offsetID, chatID); err != nil {
				return err
			}
			return nil

		default:
			log.Warn().
				Str("type", fmt.Sprintf("%T", history)).
				Msg("Unexpected message history type")
			return nil
		}
	}
}

func (f *Fetcher) processMessagesBatch(ctx context.Context, messages []tg.MessageClass, offsetID *int, chatID int) error {
	for _, msg := range messages {
		m, ok := msg.(*tg.Message)
		if !ok {
			log.Warn().
				Str("message_type", fmt.Sprintf("%T", msg)).
				Msg("Unsupported message type")
			continue
		}
		log.Info().
			Int("message_id", m.ID).
			Str("content", m.Message).
			Msg("Processing message")

		if m.Media != nil {
			if err := f.downloader.ProcessMedia(ctx, m.ID, m.Media, chatID); err != nil {
				log.Error().
					Err(err).
					Int("message_id", m.ID).
					Msg("Failed to process media for message")
			}
		}

		if *offsetID == 0 || m.ID < *offsetID {
			*offsetID = m.ID
		}
	}
	return nil
}
