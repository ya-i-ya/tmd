package internal

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tgerr"
	"github.com/sirupsen/logrus"

	"tmd/config"
)

func EnsureAuth(ctx context.Context, client *telegram.Client, cfg *config.Config) error {
	status, err := client.Auth().Status(ctx)
	if err != nil {
		return fmt.Errorf("failed to get auth status: %w", err)
	}

	if status.Authorized {
		logrus.Info("User is already authorized; no further action needed.")
		return nil
	}

	phoneNumber := cfg.Telegram.PhoneNumber
	if phoneNumber == "" {
		return errors.New("phone number is not set in config")
	}

	sentCode, err := client.Auth().SendCode(ctx, phoneNumber, auth.SendCodeOptions{
		AllowFlashCall: true,
		CurrentNumber:  true,
		AllowAppHash:   true,
	})

	if err != nil {
		var rpcErr *tgerr.Error
		if errors.As(err, &rpcErr) {
			switch rpcErr.Type {
			case "PHONE_NUMBER_INVALID":
				return fmt.Errorf("the phone number %q is invalid", phoneNumber)
			case "PHONE_NUMBER_FLOOD":
				return errors.New("too many attempts, please wait before trying again")
			case "PHONE_NUMBER_BANNED":
				return fmt.Errorf("the phone number %q is banned from Telegram", phoneNumber)
			case "PHONE_NUMBER_OCCUPIED":
				return fmt.Errorf("unexpected phone_number_occupied error: %w", err)
			default:
				return fmt.Errorf("failed to send code (rpc error: %s): %w", rpcErr.Type, err)
			}
		}
		return fmt.Errorf("failed to send code: %w", err)
	}

	logrus.Infof("A code was sent to phone number: %s", phoneNumber)

	code, err := promptInput("Enter the code you received from Telegram: ")
	if err != nil {
		return err
	}
	if code == "" {
		return errors.New("the code cannot be empty")
	}

	_, signInErr := client.Auth().SignIn(ctx, phoneNumber, sentCode.String(), code)
	if signInErr != nil {
		if errors.Is(auth.ErrPasswordAuthNeeded, signInErr) {
			if cfg.Telegram.Password == "" {
				return errors.New("this account requires a 2FA password, but config is empty")
			}
			if _, passErr := client.Auth().Password(ctx, cfg.Telegram.Password); passErr != nil {
				return fmt.Errorf("failed to authenticate with 2FA password: %w", passErr)
			}
			logrus.Info("Successfully authenticated with 2FA password.")
			return nil
		}

		var rpcErr *tgerr.Error

		if errors.As(signInErr, &rpcErr) {
			switch rpcErr.Type {
			case "PHONE_CODE_INVALID":
				return errors.New("the verification code you entered is invalid")
			case "PHONE_CODE_EXPIRED":
				return errors.New("the verification code has expired; request a new one")
			case "PHONE_NUMBER_UNOCCUPIED":

				return errors.New("this phone number is not registered on Telegram; sign-up is required")
			default:
				return fmt.Errorf("failed to sign in (rpc error: %s): %w", rpcErr.Type, signInErr)
			}
		}
		return fmt.Errorf("failed to sign in with code: %w", signInErr)
	}
	logrus.Info("Successfully authenticated with phone code.")
	return nil
}

func promptInput(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read user input: %w", err)
	}
	return strings.TrimSpace(input), nil
}
