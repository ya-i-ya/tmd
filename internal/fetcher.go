package internal

import (
	"context"
	"fmt"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/sirupsen/logrus"
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
					logrus.WithError(err).Warn("failed to process dialog")
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
					logrus.WithError(err).Warn("failed to process dialog")
				}
			}
			return nil

		default:
			logrus.Warn("unexpected dialog type")
			return nil
		}
	}
}

func (fetch *Fetcher) processDialog(ctx context.Context, dialog tg.DialogClass, users []tg.UserClass) error {
	d, ok := dialog.(*tg.Dialog)
	if !ok {
		logrus.Warnf("skipping unsupported dialog type: %T", dialog)
		return nil
	}

	switch peer := d.Peer.(type) {
	case *tg.PeerUser:
		user, found := findUser(users, peer.UserID)
		if !found {
			logrus.Warnf("user not found for peer ID: %d", peer.UserID)
			return nil
		}
		inputPeer := &tg.InputPeerUser{
			UserID:     user.ID,
			AccessHash: user.AccessHash,
		}
		return fetch.FetchAndProcessMessages(ctx, inputPeer)

	default:
		logrus.Warnf("unsupported peer type: %T", peer)
		return nil
	}
}

func (fetch *Fetcher) getNextOffsetPeer(dialog tg.DialogClass) tg.InputPeerClass {
	d, ok := dialog.(*tg.Dialog)
	if !ok {
		logrus.Warnf("unexpected dialog type when getting next offset peer: %T", dialog)
		return &tg.InputPeerEmpty{}
	}

	peer := d.GetPeer()
	if peer == nil {
		logrus.Warn("peer is nil, setting offsetPeer to InputPeerEmpty")
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
			logrus.Warnf("unexpected message history type: %T", msgs)
			return nil
		}
	}
}

func (fetch *Fetcher) processMessagesBatch(ctx context.Context, messages []tg.MessageClass, offsetID *int) error {
	for _, msg := range messages {
		m, ok := msg.(*tg.Message)
		if !ok {
			logrus.Warnf("unsupported message type: %T", msg)
			continue
		}
		logrus.Infof("Message ID: %d, Content: %s", m.ID, m.Message)

		if m.Media != nil {
			if err := fetch.downloader.ProcessMedia(ctx, m.ID, m.Media); err != nil {
				logrus.WithError(err).Errorf("Failed to process media for message ID: %d", m.ID)
			}
		}

		if *offsetID == 0 || m.ID < *offsetID {
			*offsetID = m.ID
		}
	}
	return nil
}
