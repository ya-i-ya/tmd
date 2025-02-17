package fetcher

import (
	"context"
	"errors"
	"fmt"
	"tmd/internal/db"

	"github.com/gotd/td/tg"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
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

	tgClient := tg.NewClient(f.client)

	switch peer := d.Peer.(type) {
	case *tg.PeerUser:
		if peer.UserID == f.myUserID {
			log.Debug().Msg("Skipping self-dialog")
			return nil
		}

		user, found := findUser(users, peer.UserID)
		if !found {
			inputUser := &tg.InputUser{
				UserID:     peer.UserID,
				AccessHash: 0,
			}
			users, err := tgClient.UsersGetUsers(ctx, []tg.InputUserClass{inputUser})
			if err != nil || len(users) == 0 {
				log.Warn().
					Int64("user_id", peer.UserID).
					Msg("Failed to fetch user info")
				return nil
			}

			userObj, ok := users[0].(*tg.User)
			if !ok {
				return fmt.Errorf("unexpected user type %T", users[0])
			}
			user = userObj
		}

		dialogName := user.Username
		if dialogName == "" {
			dialogName = fmt.Sprintf("user%d", user.ID)
		}

		inputPeer := &tg.InputPeerUser{
			UserID:     user.ID,
			AccessHash: user.AccessHash,
		}
		var chat db.Chat
		err := f.database.Conn.Where("telegram_id = ?", user.ID).First(&chat).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				chat = db.Chat{
					TelegramID: user.ID,
					Title:      dialogName,
				}
				if err := f.database.Conn.Create(&chat).Error; err != nil {
					return fmt.Errorf("failed to create chat record for DM: %w", err)
				}
				log.Info().Int64("chat_id", user.ID).Msg("Created new DM chat record")
			} else {
				return fmt.Errorf("failed to query DM chat record: %w", err)
			}
		} else {
			if chat.Title != dialogName {
				chat.Title = dialogName
				if err := f.database.Conn.Save(&chat).Error; err != nil {
					return fmt.Errorf("failed to update DM chat record: %w", err)
				}
			}
		}
		return f.FetchAndProcessMessages(ctx, inputPeer, dialogName, chat.ID)

	case *tg.PeerChat:
		chatID := peer.ChatID

		tgClient := tg.NewClient(f.client)
		resp, err := tgClient.MessagesGetChats(ctx, []int64{chatID})
		if err != nil {
			return fmt.Errorf("failed to get chats for chatID=%d: %w", chatID, err)
		}

		chatResp, ok := resp.(*tg.MessagesChats)
		if !ok {
			return fmt.Errorf("unexpected response type: %T", resp)
		}

		var dialogName string
		if c, ok := findChat(chatResp, chatID); ok && c.Title != "" {
			dialogName = c.Title
		} else {
			dialogName = fmt.Sprintf("chat%d", chatID)
		}

		var chat db.Chat
		err = f.database.Conn.Where("telegram_id = ?", chatID).First(&chat).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				chat = db.Chat{
					TelegramID: chatID,
					Title:      dialogName,
				}
				if err := f.database.Conn.Create(&chat).Error; err != nil {
					return fmt.Errorf("failed to create chat record: %w", err)
				}
				log.Info().Int64("chat_id", chatID).Msg("Created new chat record")
			} else {
				return fmt.Errorf("failed to query chat record: %w", err)
			}
		} else {
			if chat.Title != dialogName {
				chat.Title = dialogName
				if err := f.database.Conn.Save(&chat).Error; err != nil {
					return fmt.Errorf("failed to update chat record: %w", err)
				}
			}
		}

		inputPeer := &tg.InputPeerChat{
			ChatID: chatID,
		}
		return f.FetchAndProcessMessages(ctx, inputPeer, dialogName, chat.ID)

	default:
		log.Warn().
			Str("peer_type", fmt.Sprintf("%T", peer)).
			Msg("Unsupported peer type")
		return nil
	}
}

func findChat(res *tg.MessagesChats, chatID int64) (*tg.Chat, bool) {
	for _, c := range res.Chats {
		if chatObj, ok := c.(*tg.Chat); ok {
			if chatObj.ID == chatID {
				return chatObj, true
			}
		}
	}
	return nil, false
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
