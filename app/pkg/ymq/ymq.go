package ymq

import (
	"context"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"go.uber.org/zap"
)

type endpointResolver struct {
	endpoint string
}

func newEndpointResolver(sqsEndpoint string) *endpointResolver {
	return &endpointResolver{sqsEndpoint}
}

func (r endpointResolver) ResolveEndpoint(ctx context.Context, _ sqs.EndpointParameters) (smithyendpoints.Endpoint, error) {
	u, err := url.Parse(r.endpoint)
	if err != nil {
		return smithyendpoints.Endpoint{}, err
	}

	return smithyendpoints.Endpoint{
		URI: *u,
	}, nil
}

// Package that sets SQS endpoint to Yandex Cloud endpoint for Yandex Message Queue.
func New(ctx context.Context, accessKeyId, secretAccessKey string, sqsEndpoint string, logger *zap.Logger) (*sqs.Client, error) {
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion("ru-central1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, secretAccessKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %v", err)
	}

	return sqs.NewFromConfig(cfg, sqs.WithEndpointResolverV2(
		newEndpointResolver(sqsEndpoint),
	)), nil
}
