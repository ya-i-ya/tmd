package internal

import (
	"context"
	"fmt"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/rs/zerolog/log"
)

const (
	DialogsLimit  = 100
	MessagesLimit = 50
)

type Fetcher struct {
	client        *telegram.Client
	downloader    *Downloader
	dialogsLimit  int
	messagesLimit int
}

func NewFetcher(client *telegram.Client, downloader *Downloader, dialogsLimit, messagesLimit int) *Fetcher {
	return &Fetcher{
		client:        client,
		downloader:    downloader,
		dialogsLimit:  dialogsLimit,
		messagesLimit: messagesLimit,
	}
}

func (fetch *Fetcher) FetchAllDMs(ctx context.Context) error {
	tgClient := tg.NewClient(fetch.client)
	offsetDate, offsetID := 0, 0
	var offsetPeer tg.InputPeerClass = &tg.InputPeerEmpty{}

	for {
		res, err := tgClient.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
			OffsetDate: offsetDate,
			OffsetID:   offsetID,
			OffsetPeer: offsetPeer,
			Limit:      fetch.dialogsLimit,
			Hash:       0,
		})
		if err != nil {
			return fmt.Errorf("failed to get dialogs: %w", err)
		}

		switch d := res.(type) {
		case *tg.MessagesDialogsSlice:
			for _, dialog := range d.Dialogs {
				if err := fetch.processDialog(ctx, dialog, d.Users); err != nil {
					log.Warn().
						Err(err).
						Msg("Failed to process dialog")
				}
			}
			if len(d.Dialogs) < fetch.dialogsLimit {
				return nil
			}
			last := d.Dialogs[len(d.Dialogs)-1]
			offsetPeer = fetch.getNextOffsetPeer(last)
			offsetID = 0
			offsetDate = 0
			// Сделай НОРМАЛЬНО

		case *tg.MessagesDialogs:
			for _, dialog := range d.Dialogs {
				if err := fetch.processDialog(ctx, dialog, d.Users); err != nil {
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

func (fetch *Fetcher) processDialog(ctx context.Context, dialog tg.DialogClass, users []tg.UserClass) error {
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
		return fetch.FetchAndProcessMessages(ctx, inputPeer)

	default:
		log.Warn().
			Str("peer_type", fmt.Sprintf("%T", peer)).
			Msg("Unsupported peer type")
		return nil
	}
}

func (fetch *Fetcher) getNextOffsetPeer(dialog tg.DialogClass) tg.InputPeerClass {
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

func (fetch *Fetcher) FetchAndProcessMessages(ctx context.Context, peer tg.InputPeerClass) error {
	tgClient := tg.NewClient(fetch.client)
	offsetID := 0

	for {
		history, err := tgClient.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
			Peer:     peer,
			OffsetID: offsetID,
			Limit:    MessagesLimit,
		})
		if err != nil {
			return fmt.Errorf("failed to fetch message history: %w", err)
		}

		switch msgs := history.(type) {
		case *tg.MessagesChannelMessages:
			if len(msgs.Messages) == 0 {
				return nil
			}
			if err := fetch.processMessagesBatch(ctx, msgs.Messages, &offsetID); err != nil {
				return err
			}
			if len(msgs.Messages) < MessagesLimit {
				return nil
			}

		case *tg.MessagesMessagesSlice:
			if len(msgs.Messages) == 0 {
				return nil
			}
			if err := fetch.processMessagesBatch(ctx, msgs.Messages, &offsetID); err != nil {
				return err
			}
			if len(msgs.Messages) < MessagesLimit {
				return nil
			}

		case *tg.MessagesMessages:
			if len(msgs.Messages) == 0 {
				return nil
			}
			if err := fetch.processMessagesBatch(ctx, msgs.Messages, &offsetID); err != nil {
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

func (fetch *Fetcher) processMessagesBatch(ctx context.Context, messages []tg.MessageClass, offsetID *int) error {
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
			if err := fetch.downloader.ProcessMedia(ctx, int(m.ID), m.Media); err != nil {
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
