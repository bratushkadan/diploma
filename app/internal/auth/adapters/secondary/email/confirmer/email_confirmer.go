package email_confirmer

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/bratushkadan/floral/internal/auth/core/domain"
	"github.com/bratushkadan/floral/pkg/email"
)

type ConfirmationUrlResolver = func(ctx context.Context) (*url.URL, error)

type confirmationEmailBodyCreator struct {
	resolver ConfirmationUrlResolver
}

func (c confirmationEmailBodyCreator) Body(ctx context.Context, token string) (string, error) {
	url, err := c.resolver(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to resolve email confirmation url for email body creator: %v", err)
	}

	fmt.Printf("confirmationEmailBodyCreator: url=%s, token=%s\n", url.String(), token)

	q := url.Query()
	q.Add("token", token)
	url.RawQuery = q.Encode()

	fmt.Printf("confirmationEmailBodyCreator url with added token: url=%s\n", url.String())

	return fmt.Sprintf("Follow the link to confirm the email address: %s", url.String()), nil
}

type Email struct {
	ConfirmationSendTimeout time.Duration

	p  *email.YandexMailProvider
	bc confirmationEmailBodyCreator
}

var _ domain.EmailConfirmationSender = (*Email)(nil)

type EmailBuilder struct {
	e *Email

	senderEmail    string
	senderPassword string

	confirmationEndpoint  *string
	staticConfirmationUrl *string
}

func NewBuilder() *EmailBuilder {
	return &EmailBuilder{
		e: &Email{},
	}
}

func (b *EmailBuilder) SenderEmail(email string) *EmailBuilder {
	b.senderEmail = email
	return b
}
func (b *EmailBuilder) SenderPassword(password string) *EmailBuilder {
	b.senderPassword = password
	return b
}

// Set custom confirmation endpoint resolver.
func (b *EmailBuilder) ConfirmationEndpointResolver(r ConfirmationUrlResolver) *EmailBuilder {
	b.e.bc.resolver = r
	return b
}

// Set static url to send email confirmation to.
func (b *EmailBuilder) StaticConfirmationUrl(url string) *EmailBuilder {
	b.staticConfirmationUrl = &url
	return b
}

// Set endpoint to append to the host resolved from ctx.
func (b *EmailBuilder) ConfirmationHostCtxResolver(endpoint string) *EmailBuilder {
	b.confirmationEndpoint = &endpoint
	return b
}

// Set email confirmation sending timeout.
// 5 seconds by default.
func (b *EmailBuilder) ConfirmationSendTimeout(d time.Duration) *EmailBuilder {
	b.e.ConfirmationSendTimeout = d
	return b
}

func (b *EmailBuilder) Build() (*Email, error) {
	if b.staticConfirmationUrl != nil {
		u, err := url.Parse(*b.staticConfirmationUrl)
		if err != nil {
			return nil, fmt.Errorf("failed to parse static confirmation url: %v", err)
		}
		b.e.bc.resolver = func(_ context.Context) (*url.URL, error) {
			uCopy := *u
			return &uCopy, nil
		}
	}

	// Resolver is not yet set
	if b.e.bc.resolver == nil {
		if b.confirmationEndpoint == nil {
			return nil, errors.New("either a confirmation endpoint resolver, a confirmation endpoint or a static confirmation url must be set")
		}
		b.e.bc.resolver = newEmailConfirmationEndpointResolverCtx(*b.confirmationEndpoint)
	}

	b.e.p = email.NewYandexMailProvider(b.senderEmail, b.senderPassword)
	if b.e.ConfirmationSendTimeout == 0 {
		b.e.ConfirmationSendTimeout = 5 * time.Second
	}

	return b.e, nil
}

func (s Email) Send(ctx context.Context, in domain.EmailConfirmationSenderSendDTOInput) error {
	messageBody, err := s.bc.Body(ctx, in.ConfirmationToken)
	if err != nil {
		return fmt.Errorf("failed to create message body: %v", err)
	}

	ctx, cancel := context.WithTimeout(ctx, s.ConfirmationSendTimeout)
	defer cancel()
	return s.p.SendMail(ctx, email.EmailContents{
		To:      in.RecipientEmail,
		Subject: "Email confirmation",
		Body:    messageBody,
	})
}

func newEmailConfirmationEndpointResolverCtx(endpoint string) func(ctx context.Context) (*url.URL, error) {
	return func(ctx context.Context) (*url.URL, error) {
		host, ok := emailConfirmationHostFromContext(ctx)
		if !ok {
			return nil, errors.New("email confirmation api endpoint resolver context must have host value in order to set confirmation email")
		}

		u, err := url.Parse((&url.URL{Scheme: "https", Host: host, Path: endpoint}).String())
		if err != nil {
			return nil, fmt.Errorf("failed to parse url for email confirmation api endpoint resolver: %v", err)
		}

		return u, nil
	}

}

type keyEmailConfirmationHost int

var keyHost keyEmailConfirmationHost

// Function to inject host to the context.
// Used by the Email Confirmation Endpoint Resolver
// to resolve the host to send confirmation emails to.
// Resolver to use with the ConfirmationHostCtxResolver option.
func ContextWithEmailConfirmationHost(ctx context.Context, host string) context.Context {
	return context.WithValue(ctx, keyHost, host)
}

func emailConfirmationHostFromContext(ctx context.Context) (string, bool) {
	host, ok := ctx.Value(keyHost).(string)
	return host, ok
}
