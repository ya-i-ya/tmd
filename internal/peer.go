package internal

import (
	"context"
	"fmt"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

func ResolveChatPeer(ctx context.Context, client *telegram.Client, username string) (tg.InputPeerClass, error) {
	tgClient := tg.NewClient(client)

	resolveResult, err := tgClient.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
		Username: username,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to resolve username %q: %w", username, err)
	}

	if len(resolveResult.Chats) > 0 {
		switch ch := resolveResult.Chats[0].(type) {
		case *tg.Channel:
			return &tg.InputPeerChannel{
				ChannelID:  ch.ID,
				AccessHash: ch.AccessHash,
			}, nil
		case *tg.Chat:
			return &tg.InputPeerChat{
				ChatID: ch.ID,
			}, nil
		default:
			return nil, fmt.Errorf("unexpected chat type: %T", ch)
		}
	}

	if len(resolveResult.Users) > 0 {
		switch userObj := resolveResult.Users[0].(type) {
		case *tg.User:
			return &tg.InputPeerUser{
				UserID:     userObj.ID,
				AccessHash: userObj.AccessHash,
			}, nil
		default:
			return nil, fmt.Errorf("unexpected user type: %T", userObj)
		}
	}
	return nil, fmt.Errorf("no suitable user or channel found for username %q", username)
}
