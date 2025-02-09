package account_creation_daemon_adapter

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/bratushkadan/floral/internal/auth/core/domain"
	"github.com/bratushkadan/floral/pkg/shared/api"
	"github.com/bratushkadan/floral/pkg/sqs/rcvproc"
)

type AccountCreation struct {
	svc     domain.AuthService
	rcvProc *rcvproc.RcvProcessor[api.AccountConfirmationMessage]

	sqsQueueUrl string

	l *zap.Logger
}

type AccountCreationBuilder struct {
	ac AccountCreation

	sqs *sqs.Client
}

func NewBuilder() *AccountCreationBuilder {
	b := &AccountCreationBuilder{
		ac: AccountCreation{},
	}

	return b
}

func (b *AccountCreationBuilder) AuthService(svc domain.AuthService) *AccountCreationBuilder {
	b.ac.svc = svc
	return b
}
func (b *AccountCreationBuilder) SqsClient(sqs *sqs.Client) *AccountCreationBuilder {
	b.sqs = sqs
	return b
}
func (b *AccountCreationBuilder) SqsQueueUrl(url string) *AccountCreationBuilder {
	b.ac.sqsQueueUrl = url
	return b
}
func (b *AccountCreationBuilder) Logger(logger *zap.Logger) *AccountCreationBuilder {
	b.ac.l = logger
	return b
}

func (b *AccountCreationBuilder) Build() (*AccountCreation, error) {
	proc, err := rcvproc.New(
		b.sqs,
		b.ac.sqsQueueUrl,
		rcvproc.WithJsonDecoder[api.AccountConfirmationMessage](),
		rcvproc.WithLogger[api.AccountConfirmationMessage](b.ac.l),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set up RcvProcessor for account creation daemon sqs adapter: %v", err)
	}

	b.ac.rcvProc = proc

	return &b.ac, nil
}

func (a *AccountCreation) ReceiveProcessAccountCreationMessages(ctx context.Context) error {
	proc := func(ctx context.Context, messages []api.AccountConfirmationMessage) error {
		a.l.Info("processing account creation messages", zap.Int("count", len(messages)))
		emails := make([]string, 0, len(messages))
		for _, msg := range messages {
			emails = append(emails, msg.Email)
		}

		_, err := a.svc.ActivateAccounts(ctx, domain.ActivateAccountsReq{
			Emails: emails,
		})
		if err != nil {
			return err
		}
		a.l.Info("processed account creation messages", zap.Int("count", len(messages)))
		return nil
	}

	return a.rcvProc.RcvProcess(ctx, proc)
}
