package store

import (
	"context"
	"errors"
	"io"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Pictures struct {
	s3         *s3.Client
	endpoint   string
	bucket     string
	pathPrefix string
}

type PicturesBuilder struct {
	p Pictures
}

func NewPicturesBuilder() *PicturesBuilder {
	return &PicturesBuilder{
		p: Pictures{
			endpoint:   "https://storage.yandexcloud.net",
			pathPrefix: "product-pictures",
		},
	}
}

func (b *PicturesBuilder) S3Client(s3cl *s3.Client) *PicturesBuilder {
	b.p.s3 = s3cl
	return b
}

// Example: https://storage.yandexcloud.net
func (b *PicturesBuilder) Endpoint(endpoint string) *PicturesBuilder {
	b.p.endpoint = endpoint
	return b
}

func (b *PicturesBuilder) Bucket(bucket string) *PicturesBuilder {
	b.p.bucket = bucket
	return b
}

func (b *PicturesBuilder) PathPrefix(prefix string) *PicturesBuilder {
	b.p.pathPrefix = prefix
	return b
}

func (b *PicturesBuilder) Build() (*Pictures, error) {
	if b.p.s3 == nil {
		return nil, errors.New("s3 client must be provided")
	}

	if b.p.bucket == "" {
		return nil, errors.New("s3 bucket name must be provided")
	}
	if b.p.endpoint == "" {
		return nil, errors.New("s3 endpoint must be provided")
	}

	return &b.p, nil
}

type UploadResponse struct {
	PictureUrl      string
	PutObjectOutput *s3.PutObjectOutput
}

func (p *Pictures) Upload(ctx context.Context, path string, r io.Reader) (UploadResponse, error) {
	out, err := p.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(strings.Join([]string{p.bucket, p.pathPrefix, path}, "/")),
		Body:   r,
	})
	return UploadResponse{
		PictureUrl:      strings.Join([]string{p.endpoint, p.bucket, p.pathPrefix, path}, "/"),
		PutObjectOutput: out,
	}, err
}
func (p *Pictures) Delete(ctx context.Context, path string) (*s3.DeleteObjectOutput, error) {
	log.Print("delete object" + strings.Join([]string{p.bucket, p.pathPrefix, path}, "/"))
	return p.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(strings.Join([]string{p.bucket, p.pathPrefix, path}, "/")),
	})
}
