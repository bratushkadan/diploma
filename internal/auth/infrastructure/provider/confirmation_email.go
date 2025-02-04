package provider

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/bratushkadan/floral/internal/auth/core/domain"
	"github.com/bratushkadan/floral/pkg/email"
)

type YandexMailConfirmationProvider struct {
	ymp  *email.YandexMailProvider
	opts YandexMailConfirmationOpts
}

var _ domain.ConfirmationProvider = (*YandexMailConfirmationProvider)(nil)

type YandexMailConfirmationOpts struct {
	// The endpoint to create email confirmation link from (origin + url).
	Endpoint string
	// The query parameter to add to the web url.
	TokenQueryParameter string
}

type YandexMailConfirmationProviderConf struct {
	SenderMail       string
	SenderPass       string
	ConfirmationOpts YandexMailConfirmationOpts
}

func NewYandexMailConfirmationProvider(conf YandexMailConfirmationProviderConf) *YandexMailConfirmationProvider {
	return &YandexMailConfirmationProvider{
		ymp:  email.NewYandexMailProvider(conf.SenderMail, conf.SenderPass),
		opts: conf.ConfirmationOpts,
	}
}

func (p *YandexMailConfirmationProvider) Send(ctx context.Context, emailAddr, confirmationId string) error {
	emailBody, err := p.prepareConfirmationEmailBody(confirmationId)
	if err != nil {
		return err
	}

	email := email.EmailContents{
		To:      emailAddr,
		Subject: "Email address confirmation for account",
		Body:    emailBody,
	}

	if err := p.ymp.SendMail(ctx, email); err != nil {
		return err
	}

	return nil
}
func (p *YandexMailConfirmationProvider) prepareConfirmationEmailBody(confirmationId string) (string, error) {
	u, err := url.Parse(p.opts.Endpoint)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Add(p.opts.TokenQueryParameter, confirmationId)
	u.RawQuery = q.Encode()

	body := strings.TrimSpace(fmt.Sprintf(`
To confirm your email address, follow the link below: %s
    `, u.String()))
	return body, nil
}
