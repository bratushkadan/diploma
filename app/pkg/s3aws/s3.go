package s3aws

import (
	"context"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
)

type endpointResolver struct {
	endpoint string
}

func newEndpointResolver(s3Endpoint string) *endpointResolver {
	return &endpointResolver{s3Endpoint}
}

func (r endpointResolver) ResolveEndpoint(ctx context.Context, _ s3.EndpointParameters) (smithyendpoints.Endpoint, error) {
	u, err := url.Parse(r.endpoint)
	if err != nil {
		return smithyendpoints.Endpoint{}, err
	}

	return smithyendpoints.Endpoint{
		URI: *u,
	}, nil
}

// Package that sets S3 endpoint to Yandex Cloud Object Storage endpoint.
func New(ctx context.Context, accessKeyId, secretAccessKey string) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion("ru-central1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, secretAccessKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %v", err)
	}

	return s3.NewFromConfig(cfg, s3.WithEndpointResolverV2(
		newEndpointResolver("https://storage.yandexcloud.net"),
	)), nil
}
