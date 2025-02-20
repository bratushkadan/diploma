package ydbtopic

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicoptions"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicreader"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicwriter"
)

// Creates a YDB Topic consumer.
func NewConsumer(db *ydb.Driver, topic, consumerGroup string) (*topicreader.Reader, error) {
	r, err := db.Topic().StartReader(consumerGroup, topicoptions.ReadTopic(topic))
	if err != nil {
		return nil, fmt.Errorf("failed to create new consumer %s for ydb topic %s: %w", consumerGroup, topic, err)
	}
	return r, nil
}

// Creates a YDB Topic producer.
func NewProducer(db *ydb.Driver, topic string) (*topicwriter.Writer, error) {
	w, err := db.Topic().StartWriter(topic)
	if err != nil {
		return nil, fmt.Errorf("failed to create new producer for ydb topic %s: %w", topic, err)
	}

	return w, nil
}

// Consumes YDB Topic until ctx is cancelled, error is returned or processing callback returns an error.
// TODO: fix these design flaws, the function is solely for demonstation purposes.
func Consume(ctx context.Context, r *topicreader.Reader, cb func(data []byte) error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := r.ReadMessage(ctx)
			if err != nil {
				return fmt.Errorf("failed to read a message from topic: %w", err)
			}
			d, err := io.ReadAll(msg)
			if err != nil {
				return fmt.Errorf("failed to read message's content: %w", err)
			}

			if err := cb(d); err != nil {
				return fmt.Errorf("failed to execute message processing callback: %w", err)
			}

			if err := r.Commit(msg.Context(), msg); err != nil {
				return fmt.Errorf("failed to commit a message: %w", err)
			}
		}
	}
}

// Batch consumes YDB Topic until ctx is cancelled or error is returned.
// TODO: fix these design flaws, the function is solely for demonstation purposes.
func ConsumeBatch(ctx context.Context, r *topicreader.Reader, cb func(data [][]byte) error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			batch, err := r.ReadMessagesBatch(ctx)
			if err != nil {
				return fmt.Errorf("failed to read a message batch from topic: %w", err)
			}

			var msgs = make([][]byte, 0, len(batch.Messages))
			for _, m := range batch.Messages {
				d, err := io.ReadAll(m)
				if err != nil {
					return fmt.Errorf("failed to read batch message's content: %w", err)
				}
				msgs = append(msgs, d)
			}

			if err := cb(msgs); err != nil {
				return fmt.Errorf("failed to execute batch message processing callback: %w", err)
			}

			if err := r.Commit(batch.Context(), batch); err != nil {
				return fmt.Errorf("failed to commit a message batch: %w", err)
			}
		}
	}
}

// Produces message to YDB Topic.
func Produce(ctx context.Context, w *topicwriter.Writer, msgs ...[]byte) error {
	var wmsgs = make([]topicwriter.Message, 0, len(msgs))
	for _, v := range msgs {
		wmsgs = append(wmsgs, topicwriter.Message{
			Data: bytes.NewReader(v),
		})
	}
	if err := w.Write(ctx); err != nil {
		return fmt.Errorf("failed to procude messagesto topic: %w", err)
	}
	return nil
}
