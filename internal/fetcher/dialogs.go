package fetcher

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"github.com/rs/zerolog/log"
)

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
					log.Warn().Err(err).Msg("Failed to process dialog")
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
					log.Warn().Err(err).Msg("Failed to process dialog")
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
			log.Warn().Int64("user_id", peer.UserID).Msg("User not found for peer ID")
			return nil
		}
		inputPeer := &tg.InputPeerUser{
			UserID:     user.ID,
			AccessHash: user.AccessHash,
		}
		chatID := user.ID
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
			Msg("Unexpected type in getNextOffsetPeer")
		return &tg.InputPeerEmpty{}
	}

	peer := d.GetPeer()
	if peer == nil {
		log.Warn().Msg("Peer is nil, returning InputPeerEmpty")
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
