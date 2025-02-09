package ymq_adapter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/bratushkadan/floral/internal/auth/core/domain"
	"github.com/bratushkadan/floral/pkg/shared/api"
)

type AccountCreation struct {
	Sqs         *sqs.Client
	SqsQueueUrl string
}

var _ domain.AccountCreationNotificationProvider = (*AccountCreation)(nil)

func (q *AccountCreation) Send(ctx context.Context, in domain.SendAccountCreationNotificationDTOInput) (domain.SendAccountCreationNotificationDTOOutput, error) {
	msg := api.AccountCreationMessage{
		Id:    "",
		Email: in.Email,
	}
	emailConfirmationMsg, err := json.Marshal(&msg)
	if err != nil {
		return domain.SendAccountCreationNotificationDTOOutput{}, fmt.Errorf("failed to serialize account creation message: %v", err)
	}

	_, err = q.Sqs.SendMessage(ctx, &sqs.SendMessageInput{
		MessageBody: aws.String(string(emailConfirmationMsg)),
		QueueUrl:    aws.String(q.SqsQueueUrl),
	})
	return domain.SendAccountCreationNotificationDTOOutput{}, err
}
