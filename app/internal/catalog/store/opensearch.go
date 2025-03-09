package store

import (
	"crypto/tls"
	"net/http"

	"github.com/opensearch-project/opensearch-go"
)

type OpenSearchClientBuilder struct {
	conf opensearch.Config
}

func NewOpenSearchClientBuilder() *OpenSearchClientBuilder {
	return &OpenSearchClientBuilder{}
}

func (b *OpenSearchClientBuilder) Username(username string) *OpenSearchClientBuilder {
	b.conf.Username = username
	return b
}
func (b *OpenSearchClientBuilder) Password(password string) *OpenSearchClientBuilder {
	b.conf.Password = password
	return b
}
func (b *OpenSearchClientBuilder) Addresses(addrs []string) *OpenSearchClientBuilder {
	b.conf.Addresses = addrs
	return b
}

func (b *OpenSearchClientBuilder) Build() (*opensearch.Client, error) {
	b.conf.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return opensearch.NewClient(b.conf)
}
