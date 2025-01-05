package internal

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"tmd/pkg/cfg"
	"tmd/pkg/errors"

	ghf "github.com/go-faster/errors"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/rs/zerolog/log"
)

type flowClient struct {
	client *telegram.Client
}

func (f *flowClient) SendCode(ctx context.Context, phone string, opt auth.SendCodeOptions) (tg.AuthSentCodeClass, error) {
	return f.client.Auth().SendCode(ctx, phone, opt)
}

func (f *flowClient) SignIn(ctx context.Context, phone, code, codeHash string) (*tg.AuthAuthorization, error) {
	return f.client.Auth().SignIn(ctx, phone, code, codeHash)
}

func (f *flowClient) Password(ctx context.Context, password string) (*tg.AuthAuthorization, error) {
	return f.client.Auth().Password(ctx, password)
}

func (f *flowClient) SignUp(ctx context.Context, s auth.SignUp) (*tg.AuthAuthorization, error) {
	return f.client.Auth().SignUp(ctx, s)
}

type userAuthenticator struct {
	phone    string
	password string
}

func (u *userAuthenticator) Phone(ctx context.Context) (string, error) {
	if u.phone == "" {
		return "", errors.ErrNumberNotSet
	}
	return u.phone, nil
}

func (u *userAuthenticator) Password(ctx context.Context) (string, error) {
	return u.password, nil
}

func (u *userAuthenticator) Code(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	return promptIn("Enter the code you received from Telegram: ")
}

func (u *userAuthenticator) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	return nil
}

func (u *userAuthenticator) SignUp(ctx context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, ghf.New("sign-up is not supported")
}

func promptIn(prompt string) (string, error) {
	fmt.Print(prompt)
	r := bufio.NewReader(os.Stdin)
	text, err := r.ReadString('\n')
	if err != nil {
		return "", ghf.Wrap(err, "failed to read user input")
	}
	s := strings.TrimSpace(text)
	if s == "" {
		return "", errors.ErrCodeEmpty
	}
	return s, nil
}

func EnsureAuth(ctx context.Context, client *telegram.Client, config *cfg.Config) error {
	s, err := client.Auth().Status(ctx)
	if err != nil {
		return ghf.Wrap(err, "failed to get auth status")
	}
	if s.Authorized {
		log.Info().Msg("User is already authorized; no further action needed.")
		return nil
	}
	if config.Telegram.PhoneNumber == "" {
		return errors.ErrNumberNotSet
	}
	flow := auth.NewFlow(
		&userAuthenticator{
			phone:    config.Telegram.PhoneNumber,
			password: config.Telegram.Password,
		},
		auth.SendCodeOptions{
			AllowFlashCall: true,
			CurrentNumber:  true,
			AllowAppHash:   true,
		},
	)
	f := &flowClient{client: client}
	if err := flow.Run(ctx, f); err != nil {
		return ghf.Wrap(err, "auth flow failed")
	}
	log.Info().Msg("Successfully authenticated via AuthFlow.")
	return nil
}
