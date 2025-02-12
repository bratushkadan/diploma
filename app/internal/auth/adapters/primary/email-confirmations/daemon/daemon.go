package email_confirmation_daemon_adapter

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"go.uber.org/zap"

	"github.com/bratushkadan/floral/internal/auth/core/domain"
	"github.com/bratushkadan/floral/pkg/shared/api"
	"github.com/bratushkadan/floral/pkg/sqs/rcvproc"
)

type EmailConfirmations struct {
	svc domain.AccountEmailConfirmation

	rcvProc *rcvproc.RcvProcessor[api.AccountConfirmationMessage]

	l *zap.Logger
}

type EmailConfirmationBuilder struct {
	ec EmailConfirmations

	sqsQueueUrl string
	sqs         *sqs.Client
}

func NewBuilder() *EmailConfirmationBuilder {
	b := &EmailConfirmationBuilder{
		ec: EmailConfirmations{},
	}

	return b
}

func (b *EmailConfirmationBuilder) Service(svc domain.AccountEmailConfirmation) *EmailConfirmationBuilder {
	b.ec.svc = svc
	return b
}
func (b *EmailConfirmationBuilder) SqsClient(sqs *sqs.Client) *EmailConfirmationBuilder {
	b.sqs = sqs
	return b
}
func (b *EmailConfirmationBuilder) SqsQueueUrl(url string) *EmailConfirmationBuilder {
	b.sqsQueueUrl = url
	return b
}
func (b *EmailConfirmationBuilder) Logger(logger *zap.Logger) *EmailConfirmationBuilder {
	b.ec.l = logger
	return b
}

func (b *EmailConfirmationBuilder) Build() (*EmailConfirmations, error) {
	proc, err := rcvproc.New(
		b.sqs,
		b.sqsQueueUrl,
		rcvproc.WithJsonDecoder[api.AccountConfirmationMessage](),
		rcvproc.WithLogger[api.AccountConfirmationMessage](b.ec.l),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set up RcvProcessor for account confirmation daemon sqs adapter: %v", err)
	}

	b.ec.rcvProc = proc

	return &b.ec, nil
}

func (a *EmailConfirmations) ReceiveProcessAccountCreationMessages(ctx context.Context) error {
	proc := func(ctx context.Context, messages []api.AccountConfirmationMessage) error {
		a.l.Info("processing account confirmation messages", zap.Int("count", len(messages)))
		emails := make([]string, 0, len(messages))
		for _, msg := range messages {
			emails = append(emails, msg.Email)
		}

		for _, email := range emails {
			if err := a.svc.Send(ctx, email); err != nil {
				return err
			}
		}

		a.l.Info("processed account confirmation messages", zap.Int("count", len(messages)))
		return nil
	}

	return a.rcvProc.RcvProcess(ctx, proc)
}
