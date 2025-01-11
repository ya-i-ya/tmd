package fetcher

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"github.com/rs/zerolog/log"
)

func (f *Fetcher) FetchAndProcessMessages(ctx context.Context, peer tg.InputPeerClass, chatID int64) error {
	tgClient := tg.NewClient(f.client)
	offsetID := 0

	for {
		history, err := tgClient.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
			Peer:     peer,
			OffsetID: offsetID,
			Limit:    f.messagesLimit,
		})
		if err != nil {
			return fmt.Errorf("failed to fetch message history: %w", err)
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

func (f *Fetcher) processMessagesBatch(
	ctx context.Context,
	messages []tg.MessageClass,
	offsetID *int,
	chatID int64,
) error {
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
			localPath, err := f.downloader.ProcessMedia(ctx, m.ID, m.Media, chatID)
			if err != nil {
				log.Error().
					Err(err).
					Int("message_id", m.ID).
					Msg("Failed to process media for message")
				continue
			}

			userID, ok := peerUserID(m.FromID)
			if !ok {
				userID = 0
			}

			job := MeJob{
				MessageID:      m.ID,
				TelegramUserID: userID,
				ChatID:         chatID,
				Media:          m.Media,
				MediaFilePath:  localPath,
			}
			f.meChan <- job
		}

		if *offsetID == 0 || m.ID < *offsetID {
			*offsetID = m.ID
		}
	}
	return nil
}

func peerUserID(peer tg.PeerClass) (int64, bool) {
	switch p := peer.(type) {
	case *tg.PeerUser:
		return p.UserID, true
	default:
		return 0, false
	}
}
