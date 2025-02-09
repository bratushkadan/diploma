package email_confirmation_daemon_adapter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"go.uber.org/zap"

	"github.com/bratushkadan/floral/internal/auth/core/domain"
	"github.com/bratushkadan/floral/pkg/shared/api"
	"github.com/bratushkadan/floral/pkg/ymq"
)

type EmailConfirmations struct {
	svc domain.AuthService
	ymq *ymq.Ymq
	l   *zap.Logger
}

func New(ymq *ymq.Ymq, logger *zap.Logger) *EmailConfirmations {
	return &EmailConfirmations{ymq: ymq, l: logger}
}

// FIXME: dont return error on errors on messages processing, receive chan error by an argument and return error only on fatal errors
func (q *EmailConfirmations) RcvProcess(ctx context.Context, process func(ctx context.Context, dto []api.AccountConfirmationMessage) error) error {
	// Long Polling
	// https://docs.amazonaws.cn/en_us/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-short-and-long-polling.html
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			output, err := q.ymq.Cl.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
				QueueUrl:            aws.String(q.ymq.Endpoint()),
				MaxNumberOfMessages: 10,
				WaitTimeSeconds:     5,
			})
			if err != nil {
				return fmt.Errorf("failed to receive email confirmation ymq sqs messages: %v", err)
			}

			emailConfirmations := make([]api.AccountConfirmationMessage, 0, len(output.Messages))
			deleteMessageBatchReqEntries := make([]types.DeleteMessageBatchRequestEntry, 0, len(output.Messages))
			for _, message := range output.Messages {
				var emailConfirmation api.AccountConfirmationMessage
				if err := json.Unmarshal([]byte(*message.Body), &emailConfirmation); err != nil {
					return fmt.Errorf("failed to unmarshal email confirmation ymq sqs message: %v", err)
				}
				emailConfirmations = append(emailConfirmations, emailConfirmation)
				deleteMessageBatchReqEntries = append(deleteMessageBatchReqEntries, types.DeleteMessageBatchRequestEntry{
					Id:            message.MessageId,
					ReceiptHandle: message.ReceiptHandle,
				})
			}

			if err := process(ctx, emailConfirmations); err != nil {
				return fmt.Errorf("failed to process email confirmation messages for ymq sqs: %v", err)
			}

			_, err = q.ymq.Cl.DeleteMessageBatch(ctx, &sqs.DeleteMessageBatchInput{
				QueueUrl: aws.String(q.ymq.Endpoint()),
				Entries:  deleteMessageBatchReqEntries,
			})
			if err != nil {
				return fmt.Errorf("failed to delete processed messages from email confirmation ymq sqs: %v", err)
			}
		}
	}
}
func (q *EmailConfirmations) PublishConfirmation(ctx context.Context, confirmation api.AccountConfirmationMessage) error {
	emailConfirmationMsg, err := json.Marshal(&confirmation)
	if err != nil {
		return fmt.Errorf("failed to serialize email confirmation dto: %v", err)
	}
	_, err = q.ymq.Cl.SendMessage(ctx, &sqs.SendMessageInput{
		MessageBody: aws.String(string(emailConfirmationMsg)),
		QueueUrl:    aws.String(q.ymq.Endpoint()),
	})
	return err
}
