package yds

import (
	"context"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
)

type endpointResolver struct {
	endpoint string
}

func newEndpointResolver(sqsEndpoint string) *endpointResolver {
	return &endpointResolver{sqsEndpoint}
}

func (r endpointResolver) ResolveEndpoint(ctx context.Context, _ kinesis.EndpointParameters) (smithyendpoints.Endpoint, error) {
	u, err := url.Parse(r.endpoint)
	if err != nil {
		return smithyendpoints.Endpoint{}, err
	}

	return smithyendpoints.Endpoint{
		URI: *u,
	}, nil
}

// Package that sets Kinesis endpoint to Yandex Cloud endpoint for Yandex Data Stream.
// Kinesis endpoint is "https://yds.serverless.yandexcloud.net"
func New(ctx context.Context, accessKeyId, secretAccessKey string, kinesisEndpoint string) (*kinesis.Client, error) {
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion("ru-central1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, secretAccessKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %v", err)
	}

	return kinesis.NewFromConfig(cfg, kinesis.WithEndpointResolverV2(
		newEndpointResolver(kinesisEndpoint),
	)), nil
}
