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
  INDEX idx_list_orders GLOBAL ASYNC on (user_id, created_at)
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

```sql
CREATE TABLE `orders/payments` (
  id Utf8 NOT NULL,
  order_id Utf8 NOT NULL,
  amount Double NOT NULL,
  currency_iso_4217 Uint32 NOT NULL,
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

```sh
TF_OUTPUT=$(../terraform/tf output -json -no-color)
export YDB_ENDPOINT="$(echo "${TF_OUTPUT}" | jq -cMr .ydb.value.full_endpoint)"
export YDB_SERVICE_ACCOUNT_KEY_FILE_CREDENTIALS="$(scripts/ydb_access_token.sh)"
export YDB_AUTH_METHOD=environ
APP_SA_STATIC_KEY_SECRET_ID="$(echo $TF_OUTPUT | jq -cMr .app_sa.value.static_key_lockbox_secret_id)"
SECRET=$(yc lockbox payload get "${APP_SA_STATIC_KEY_SECRET_ID}")
INFRA_TOKENS_SECRET_ID="$(echo $TF_OUTPUT | jq -cMr .infra_tokens_lockbox_secret_id.value)"
INFRA_TOKENS_SECRET="$(yc lockbox payload get "${INFRA_TOKENS_SECRET_ID}")"
export APP_AUTH_TOKEN_PUBLIC_KEY="$(echo $INFRA_TOKENS_SECRET | yq -M '.entries.[] | select(.key == "auth_token_public.key").text_value')"
YOOMONEY_NOTIFICATIONS_SECRET_SECRET_ID="$(echo $TF_OUTPUT | jq -cMr .yoomoney_payment_provider_notifications_secret_secret_id.value)"
YOOMONEY_NOTIFICATIONS_SECRET_PAYLOAD=$(yc lockbox payload get "${YOOMONEY_NOTIFICATIONS_SECRET_SECRET_ID}")
export YOOMONEY_NOTIFICATIONS_SECRET="$(echo $YOOMONEY_NOTIFICATIONS_SECRET_PAYLOAD | yq -M '.entries.[] | select(.key == "notification_secret").text_value')"
go run cmd/orders/main.go
```

## Testing

### Run fake produce Yoomoney Payment Notification

prepare:

```sh
TF_OUTPUT=$(../terraform/tf output -json -no-color)
YOOMONEY_NOTIFICATIONS_SECRET_SECRET_ID="$(echo $TF_OUTPUT | jq -cMr .yoomoney_payment_provider_notifications_secret_secret_id.value)"
YOOMONEY_NOTIFICATIONS_SECRET_PAYLOAD=$(yc lockbox payload get "${YOOMONEY_NOTIFICATIONS_SECRET_SECRET_ID}")
export YOOMONEY_NOTIFICATIONS_SECRET="$(echo $YOOMONEY_NOTIFICATIONS_SECRET_PAYLOAD | yq -M '.entries.[] | select(.key == "notification_secret").text_value')"
```

run

```sh
go run cmd/orders/tests/produce-yoomoney-payment-notification/main.go
```

## Build docker image locally

1\. `cd app`
2\. `go mod tidy`
3\. `CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin cmd/orders/main.go`

### Build for Yandex Cloud Container Registry

1. Bump `local.versions.${SERVICE}` in Terraform
2. Run the following command:

```sh
export SERVICE="orders"
export TAG="$(echo "local.versions.${SERVICE}" | ./terraform/tf console | jq -cMr)"
./app/scripts/build-push-image.sh
```