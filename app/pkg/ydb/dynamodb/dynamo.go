package ydb_dynamodb

import (
	"context"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
)

type ydbDocApiEndpointResolver struct {
	endpoint string
}

func (r ydbDocApiEndpointResolver) ResolveEndpoint(ctx context.Context, _ dynamodb.EndpointParameters) (smithyendpoints.Endpoint, error) {
	u, err := url.Parse(r.endpoint)
	if err != nil {
		return smithyendpoints.Endpoint{}, err
	}

	return smithyendpoints.Endpoint{
		URI: *u,
	}, nil
}

func newYdbDocApiEndpointResolver(endpoint string) *ydbDocApiEndpointResolver {
	return &ydbDocApiEndpointResolver{
		endpoint: endpoint,
	}
}

func New(ctx context.Context, accessKeyId, secretAccessKey string, ydbDocApiEndpoint string) (*dynamodb.Client, error) {
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion("ru-central1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, secretAccessKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config for ydb dynamodb: %v", err)
	}

	return dynamodb.NewFromConfig(cfg, dynamodb.WithEndpointResolverV2(
		newYdbDocApiEndpointResolver(ydbDocApiEndpoint),
	)), nil
}
