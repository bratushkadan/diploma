package rcvproc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type Decoder[T any] func(body string, target *T) error

type RcvProcessor[T any] struct {
	sqs         *sqs.Client
	sqsQueueUrl string

	decoder Decoder[T]
}

func jsonDecoder[T any](body string, target *T) error {
	return json.Unmarshal([]byte(body), target)
}

func stringDecoder[T any](body string, target *T) error {
	bodyAny := any(body).(T)
	*target = bodyAny
	return nil
}

type Option[T any] func(p *RcvProcessor[T]) error

func WithJsonDecoder[T any]() Option[T] {
	return WithDecoder(jsonDecoder[T])
}
func WithDecoder[T any](d Decoder[T]) Option[T] {
	return func(p *RcvProcessor[T]) error {
		p.decoder = d
		return nil
	}
}

func New[T any](sqs *sqs.Client, sqsQueueUrl string, opts ...Option[T]) (*RcvProcessor[T], error) {
	proc := RcvProcessor[T]{
		sqs:         sqs,
		sqsQueueUrl: sqsQueueUrl,
	}

	for _, opt := range opts {
		opt(&proc)
	}

	if proc.decoder == nil {
		proc.decoder = stringDecoder
	}

	return &proc, nil
}

// FIXME: dont return error on errors on messages processing, receive chan error by an argument and return error only on fatal errors
func (q *RcvProcessor[T]) RcvProcess(ctx context.Context, process func(ctx context.Context, messages []T) error) error {
	// Long Polling
	// https://docs.amazonaws.cn/en_us/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-short-and-long-polling.html
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			output, err := q.sqs.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
				QueueUrl:            aws.String(q.sqsQueueUrl),
				MaxNumberOfMessages: 10,
				WaitTimeSeconds:     5,
			})
			if err != nil {
				return fmt.Errorf("failed to receive ymq sqs messages: %v", err)
			}

			decodedMsgs := make([]T, 0, len(output.Messages))
			deleteMessageBatchReqEntries := make([]types.DeleteMessageBatchRequestEntry, 0, len(output.Messages))
			for _, message := range output.Messages {
				var decodedMsg T
				if err := json.Unmarshal([]byte(*message.Body), &decodedMsg); err != nil {
					return fmt.Errorf("failed to unmarshal ymq sqs message: %v", err)
				}
				decodedMsgs = append(decodedMsgs, decodedMsg)
				deleteMessageBatchReqEntries = append(deleteMessageBatchReqEntries, types.DeleteMessageBatchRequestEntry{
					Id:            message.MessageId,
					ReceiptHandle: message.ReceiptHandle,
				})
			}

			if err := process(ctx, decodedMsgs); err != nil {
				return fmt.Errorf("failed to process messages for ymq sqs: %v", err)
			}

			_, err = q.sqs.DeleteMessageBatch(ctx, &sqs.DeleteMessageBatchInput{
				QueueUrl: aws.String(q.sqsQueueUrl),
				Entries:  deleteMessageBatchReqEntries,
			})
			if err != nil {
				return fmt.Errorf("failed to delete processed messages from ymq sqs: %v", err)
			}
		}
	}
}
