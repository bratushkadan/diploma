# Catalog microservice

## Prepare

Create `./terraform/opensearch-creds.yaml` file and populate it with values `user` and `password`.

## Components

### OpenSearch

Ports:
- `9200` - RESTful API port.

### Run migration

1. Do the port forwarding for server.
2. Execute the command (from the app/ directory):

```sh
OPENSEARCH_USER=admin
OPENSEARCH_PASSWORD=
go run cmd/catalog/opensearch-indices/main.go
```

### Run app locally

```sh
OPENSEARCH_USER=admin
OPENSEARCH_PASSWORD=
go run cmd/catalog/main.go
```

### CURLs for testing

#### Test app locally

```sh
curl -s -H 'Content-Type: application/json' \
    http://localhost:8080/api/v1/catalog
```

```sh
curl -s -H 'Content-Type: application/json' \
    http://localhost:8080/api/v1/catalog?filter=крупа
```

```sh
сurl -XPOST \
  -H 'Content-Type: application/json' \
  -d @catalog-sync-body-example.json \
  http://localhost:8080/api/internal/v1/sync-catalog
```

### Connect to instance

```sh
HOST="$(yc compute instance get --name opensearch --format json | jq -cMr '.network_interfaces[0].primary_v4_address.one_to_one_nat.address')"
ssh "${HOST}"
```

### Port-forward services

#### OpenSearch server

```sh
HOST="$(yc compute instance get --name opensearch --format json | jq -cMr '.network_interfaces[0].primary_v4_address.one_to_one_nat.address')"
ssh -L 9200:localhost:9200 "${HOST}"
```

#### Dashboard

```sh
HOST="$(yc compute instance get --name opensearch --format json | jq -cMr '.network_interfaces[0].primary_v4_address.one_to_one_nat.address')"
ssh -L 5601:localhost:5601 "${HOST}"
```

#### Run in docker

```sh
docker-compose up -d
```

#### Open dashboard

[Dashboard in web browser](http://localhost:5601)

##### Console Queries

List products:

```sh
GET /products/_search
{
  "query": {
    "match_all": {}
  }
}
```

## Build docker image locally

1\. `cd app`
2\. `go mod tidy`
3\. `CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin cmd/auth/email-confirmation/main.go`

### Build for Yandex Cloud Container Registry

Email confirmation:

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin cmd/catalog/main.go
TAG=0.0.2
docker build -f build/catalog/Dockerfile -t "catalog:${TAG}" .
rm bin
yc iam create-token | docker login cr.yandex -u iam --password-stdin
TARGET="cr.yandex/$(../terraform/tf output -json -no-color | jq -cMr .container_registry.value.repository.catalog.name):${TAG}"
docker tag "catalog:${TAG}" "${TARGET}"
docker push "${TARGET}"
```
