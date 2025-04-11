# Order microservice

## Data Model

[DB Diagram](https://dbdiagram.io/d/ecom-67b96d09263d6cf9a01083b2)

YDB Schema:

```sql
CREATE TABLE `orders/orders` (
  id Utf8 NOT NULL,
  user_id Utf8 NOT NULL,
  -- For ease of designing and implementing business processes only online payments are allowed
  -- online_payment Bool NOT NULL DEFAULT false,
  status Utf8 NOT NULL,
  created_at Datetime NOT NULL,
  updated_at Datetime NOT NULL,
  PRIMARY KEY (id),
  INDEX idx_user_id GLOBAL ASYNC on (user_id)
);
```

```sql
CREATE TABLE `orders/order_items` (
  order_id Utf8 NOT NULL,
  product_id Utf8 NOT NULL,
  name Utf8 NOT NULL,
  seller_id Utf8 NOT NULL,
  count Uint32 NOT NULL,
  price Double NOT NULL,
  picture Utf8,
  PRIMARY KEY (order_id, product_id)
);
```

`contents` field schema:
```go
type OrderContentsItem struct {
    ProductId Utf8 `json:"product_id"`
    Name Utf8 `json:"name"`
    SellerId Utf8 `json:"seller_id"`
    Count int32 `json:"count"`
    Price float64 `json:"price"`
    // Picture url
    Picture *Utf8 `json:"picture"`
}
type OrderContents = OrderContentsItem[] 
```

```sql
CREATE TABLE `orders/payments` (
  id Utf8 NOT NULL,
  order_id Utf8 NOT NULL,
  provider Json NOT NULL,
  created_at Timestamp NOT NULL,
  updated_at Timestamp NOT NULL,
  refunded_at Timestamp,
  PRIMARY KEY (id),
  INDEX idx_order_id GLOBAL ASYNC on (order_id)
);
```

```sql
CREATE TABLE `orders/operations` (
  id Utf8 NOT NULL,
  type Utf8 NOT NULL,
  status Utf8 NOT NULL,
  details Utf8,
  user_id Utf8 NOT NULL,
  order_id Utf8,
  created_at Timestamp NOT NULL,
  updated_at Timestamp NOT NULL,
  PRIMARY KEY (id),
  INDEX idx_status GLOBAL ASYNC on (status),
  INDEX idx_order_id GLOBAL ASYNC on (order_id)
);
```

## Private endpoints

- Process cart contents ("cart contents" event/message)
- Process reserved products contents ("reserved products contents" event/message)
- Cancel unpaid orders (invoked by *Timer* Serverless Trigger)
- Update order in `cancelling` status ("unreserved products for order" event/message)

## General idea

Orders that are older than one hour and are not paid online (if not paid by cash) are cancelled. Order cancellation is scheduled regularly.

## Run

### Setup env and run

## CURLs for testing

## Build docker image locally

1\. `cd app`
2\. `go mod tidy`
3\. `CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin cmd/orders/main.go`

### Build for Yandex Cloud Container Registry

Email confirmation:

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin cmd/orders/main.go
TAG=0.0.1
docker build -f build/orders/Dockerfile -t "orders:${TAG}" .
rm bin
yc iam create-token | docker login cr.yandex -u iam --password-stdin
TARGET="cr.yandex/$(../terraform/tf output -json -no-color | jq -cMr .container_registry.value.repository.orders.name):${TAG}"
docker tag "orders:${TAG}" "${TARGET}"
docker push "${TARGET}"
```