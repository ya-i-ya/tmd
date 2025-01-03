package err

import (
	"errors"
	"fmt"
	"github.com/gotd/td/telegram/auth"

	"github.com/gotd/td/tgerr"
)

var (
	ErrNumberNotSet  = errors.New("phone number is not set in cfg")
	ErrCodeEmpty     = errors.New("the verification code cannot be empty")
	ErrPasswordEmpty = errors.New("this account requires a 2FA password, but cfg is empty")
)

func HandleTGError(err error) error {
	var rpcErr *tgerr.Error
	if errors.As(err, &rpcErr) {
		switch rpcErr.Type {
		case "PHONE_NUMBER_INVALID":
			return fmt.Errorf("the phone number is invalid")
		case "PHONE_NUMBER_FLOOD":
			return fmt.Errorf("too many attempts, please wait before trying again")
		case "PHONE_NUMBER_BANNED":
			return fmt.Errorf("the phone number is banned from Telegram")
		case "PHONE_CODE_INVALID":
			return fmt.Errorf("the verification code you entered is invalid")
		case "PHONE_CODE_EXPIRED":
			return fmt.Errorf("the verification code has expired; request a new one")
		case "PHONE_NUMBER_UNOCCUPIED":
			return fmt.Errorf("this phone number is not registered on Telegram; sign-up is required")
		default:
			return fmt.Errorf("telegram RPC error: %s", rpcErr.Type)
		}
	}
	return nil
}

func Is2FAError(err error) bool {
	return errors.Is(err, auth.ErrPasswordAuthNeeded)
}
