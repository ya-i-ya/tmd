package internal

import (
	"context"
	"errors"
	"fmt"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"tmd/config"
)

func EnsureAuth(ctx context.Context, client *telegram.Client, cfg *config.Config) error {
	status, err := client.Auth().Status(ctx)
	if err != nil {
		return fmt.Errorf("failed to get auth status: %w", err)
	}
	if status.Authorized {
		return nil
	}

	phone := cfg.Telegram.PhoneNumber

	sentCode, err := client.Auth().SendCode(ctx, phone, auth.SendCodeOptions{
		AllowFlashCall: true,
		CurrentNumber:  true,
		AllowAppHash:   true,
	})
	if err != nil {
		return fmt.Errorf("failed to send code: %w", err)
	}

	fmt.Print("Enter the code from Telegram: ")
	var userInputCode string
	_, err = fmt.Scanln(&userInputCode)
	if err != nil {
		return fmt.Errorf("failed to read user input: %w", err)
	}

	if userInputCode == "" {
		return errors.New("the code cannot be empty")
	}

	_, err = client.Auth().SignIn(ctx, phone, sentCode.String(), userInputCode)
	if err != nil {
		return err
	}
	return nil
}
