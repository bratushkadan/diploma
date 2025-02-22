package ydb_dynamodb_adapter

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/bratushkadan/floral/internal/auth/core/domain"
	ydb_dynamodb "github.com/bratushkadan/floral/pkg/ydb/dynamodb"
	"go.uber.org/zap"
)

const (
	tableEmailConfirmationTokens = "auth/email_confirmation_tokens"
)

var _ domain.EmailConfirmationTokens = (*EmailConfirmationTokens)(nil)

type EmailConfirmationTokens struct {
	cl *dynamodb.Client
	l  *zap.Logger
}

func NewEmailConfirmationTokens(ctx context.Context, accessKeyId, secretAccessKey string, ydbDocApiEndpoint string, logger *zap.Logger) (*EmailConfirmationTokens, error) {
	client, err := ydb_dynamodb.New(ctx, accessKeyId, secretAccessKey, ydbDocApiEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to setup dynamodb email confirmator: %v", err)
	}
	return &EmailConfirmationTokens{cl: client, l: logger}, nil
}

func (db *EmailConfirmationTokens) InsertToken(ctx context.Context, email, token string) error {
	item := &dynamodb.PutItemInput{
		TableName: aws.String(tableEmailConfirmationTokens),
		Item: map[string]types.AttributeValue{
			"email":      &types.AttributeValueMemberS{Value: email},
			"token":      &types.AttributeValueMemberS{Value: token},
			"expires_at": &types.AttributeValueMemberN{Value: strconv.FormatInt(time.Now().Add(20*time.Minute).Unix(), 10)},
		},
	}

	output, err := db.cl.PutItem(ctx, item)
	if err != nil {
		return nil
	}

	var unmarshaledItem domain.EmailConfirmationRecord
	if err := attributevalue.UnmarshalMap(output.Attributes, &unmarshaledItem); err != nil {
		return err
	}

	db.l.Info("inserted email token", zap.String("email", email))
	return nil
}

func (db *EmailConfirmationTokens) ListTokensEmail(ctx context.Context, email string) ([]domain.EmailConfirmationRecord, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(tableEmailConfirmationTokens),
		KeyConditionExpression: aws.String("email = :emailVal"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":emailVal": &types.AttributeValueMemberS{Value: email},
		},
	}

	result, err := db.cl.Query(ctx, input)
	if err != nil {
		return nil, err
	}

	var tokenRecords []domain.EmailConfirmationRecord
	for _, item := range result.Items {
		var unmarshaledItem domain.EmailConfirmationRecord
		if err := attributevalue.UnmarshalMap(item, &unmarshaledItem); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response from dynamodb: %v", err)
		}
		tokenRecords = append(tokenRecords, unmarshaledItem)
	}

	return tokenRecords, nil
}
func (db *EmailConfirmationTokens) FindTokenRecord(ctx context.Context, token string) (*domain.EmailConfirmationRecord, error) {
	filtEx := expression.Name("token").Equal(expression.Value(token))
	projEx := expression.NamesList(
		expression.Name("email"),
		expression.Name("token"),
		expression.Name("expires_at"),
	)

	expr, err := expression.NewBuilder().
		WithFilter(filtEx).
		WithProjection(projEx).
		Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build find token record query: %v", err)
	}

	input := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		TableName:                 aws.String(tableEmailConfirmationTokens),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
	}
	result, err := db.cl.Scan(ctx, input)
	if err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, nil
	}

	var unmarshaledItem domain.EmailConfirmationRecord
	if err := attributevalue.UnmarshalMap(result.Items[0], &unmarshaledItem); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response from dynamodb: %v", err)
	}
	return &unmarshaledItem, nil
}
