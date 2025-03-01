package fetcher

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"tmd/internal/db"

	"github.com/gotd/td/tg"
	"github.com/rs/zerolog/log"
)

func (f *Fetcher) FetchAndProcessMessages(ctx context.Context, peer tg.InputPeerClass, dialogName string, chatUUID uuid.UUID) error {
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
			if err := f.processMessagesBatch(ctx, msgs.Messages, &offsetID, dialogName, chatUUID); err != nil {
				return err
			}
			if len(msgs.Messages) < f.messagesLimit {
				return nil
			}

		case *tg.MessagesMessagesSlice:
			if len(msgs.Messages) == 0 {
				return nil
			}
			if err := f.processMessagesBatch(ctx, msgs.Messages, &offsetID, dialogName, chatUUID); err != nil {
				return err
			}
			if len(msgs.Messages) < f.messagesLimit {
				return nil
			}

		case *tg.MessagesMessages:
			if len(msgs.Messages) == 0 {
				return nil
			}
			if err := f.processMessagesBatch(ctx, msgs.Messages, &offsetID, dialogName, chatUUID); err != nil {
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
	dialogName string,
	chatUUID uuid.UUID,
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

		senderUserID, valid := peerUserID(m.FromID)
		if !valid || senderUserID == 0 {
			log.Warn().Int("message_id", m.ID).Msg("Skipping message with invalid sender")
			continue
		}

		var userRecord db.User
		if senderUserID != 0 {
			err := f.database.Conn.Where("telegram_user_id = ?", senderUserID).
				First(&userRecord).
				Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					userRecord = db.User{
						TelegramUserID: senderUserID,
						Username:       fmt.Sprintf("user%d", senderUserID),
					}
					if errCreate := f.database.Conn.Create(&userRecord).Error; errCreate != nil {
						log.Error().Err(errCreate).Msg("Failed to create user record")
					}
				} else {
					log.Error().Err(err).Msg("Failed to query user record")
				}
			}
		}

		msgType := "text"
		if m.Media != nil {
			switch m.Media.(type) {
			case *tg.MessageMediaPhoto:
				msgType = "photo"
			case *tg.MessageMediaDocument:
				msgType = "document"
			}
		}

		messageRecord := db.Message{
			MessageID:   m.ID,
			ChatID:      chatUUID,
			UserID:      userRecord.ID,
			Content:     m.Message,
			MessageType: msgType,
		}

		if err := f.database.Conn.
			Where("message_id = ? AND chat_id = ?", m.ID, chatUUID).
			FirstOrCreate(&messageRecord).Error; err != nil {
			log.Error().Err(err).Msg("Failed to create/find message record")
		}

		if m.Media != nil {
			job := MeJob{
				MessageID:      m.ID,
				TelegramUserID: senderUserID,
				Media:          m.Media,
				DialogName:     dialogName,
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
