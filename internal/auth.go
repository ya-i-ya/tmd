package internal

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	stdErrors "errors"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/rs/zerolog/log"
	"tmd/cfg"
	"tmd/errors"
)

func EnsureAuth(ctx context.Context, client *telegram.Client, cfg *cfg.Config) error {
	status, err := client.Auth().Status(ctx)
	if err != nil {
		return fmt.Errorf("failed to get auth status: %w", err)
	}

	if status.Authorized {
		log.Info().Msg("User is already authorized; no further action needed.")
		return nil
	}

	phoneNumber := cfg.Telegram.PhoneNumber
	if phoneNumber == "" {
		return errors.ErrNumberNotSet
	}

	sentCode, err := client.Auth().SendCode(ctx, phoneNumber, auth.SendCodeOptions{
		AllowFlashCall: true,
		CurrentNumber:  true,
		AllowAppHash:   true,
	})
	if err != nil {
		if tgErr := errors.HandleTGError(err); tgErr != nil {
			return tgErr
		}
		return fmt.Errorf("failed to send code: %w", err)
	}

	if sentCode.String() == "" {
		log.Error().Msg("PhoneCodeHash is empty in SentCode response")
		return stdErrors.New("PhoneCodeHash is empty in the SentCode response")
	}

	code, err := promptInput("Enter the code you received from Telegram: ")
	if err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	if _, err := client.Auth().SignIn(ctx, phoneNumber, sentCode.String(), code); err != nil {
		if errors.Is2FAError(err) {
			return handleTwoFactorAuth(ctx, client, cfg)
		}
		if tgErr := errors.HandleTGError(err); tgErr != nil {
			return tgErr
		}
		return fmt.Errorf("failed to sign in with the provided code: %w", err)
	}

	log.Info().Msg("Successfully authenticated with phone code.")
	return nil

}

func handleTwoFactorAuth(ctx context.Context, client *telegram.Client, cfg *cfg.Config) error {
	password := cfg.Telegram.Password
	if password == "" {
		return errors.ErrPasswordEmpty
	}

	if _, err := client.Auth().Password(ctx, password); err != nil {
		return fmt.Errorf("failed to authenticate with 2FA password: %w", err)
	}

	log.Info().Msg("Successfully authenticated with 2FA password.")
	return nil
}

func promptInput(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read user input: %w", err)
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return "", errors.ErrCodeEmpty
	}
	return input, nil
}
