# Cart microservice

## Data Model

[DB Diagram](https://dbdiagram.io/d/ecom-67b96d09263d6cf9a01083b2)

YDB Schema:

```sql
CREATE TABLE `cart/positions` (
    user_id Utf8 NOT NULL,
    product_id Utf8 NOT NULL,
    count Uint32 NOT NULL,
    PRIMARY KEY (user_id, product_id),
);
```

## SEED(s) use cases

- Add product to cart (or change count of products in cart)
- Delete product from cart

## Private endpoints

- Process publish cart contents (process event/message)
- Process clear cart (process event/message)

## Details

No more than 25 distinct items in cart.

If a user has products from one seller in their cart and a product from another seller is added to the cart, cart is first cleared and then product from another seller is added.

## Run

### Setup env and run

```sh
TF_OUTPUT=$(../terraform/tf output -json -no-color)
export YDB_ENDPOINT="$(echo "${TF_OUTPUT}" | jq -cMr .ydb.value.full_endpoint)"
export YDB_SERVICE_ACCOUNT_KEY_FILE_CREDENTIALS="$(scripts/ydb_access_token.sh)"
export YDB_AUTH_METHOD=environ
INFRA_TOKENS_SECRET_ID="$(echo $TF_OUTPUT | jq -cMr .infra_tokens_lockbox_secret_id.value)"
INFRA_TOKENS_SECRET="$(yc lockbox payload get "${INFRA_TOKENS_SECRET_ID}")"
export APP_AUTH_TOKEN_PUBLIC_KEY="$(echo $INFRA_TOKENS_SECRET | yq -M '.entries.[] | select(.key == "auth_token_public.key").text_value')"
go run cmd/cart/main.go
```

## CURLs for testing

## Build docker image locally

1\. `cd app`
2\. `go mod tidy`
3\. `CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin cmd/cart/main.go`

### Build for Yandex Cloud Container Registry

Email confirmation:

1. Bump `local.versions.${SERVICE}` in Terraform
2. Run the following command:

```sh
export SERVICE="cart"
export TAG="$(echo "local.versions.${SERVICE}" | ./terraform/tf console | jq -cMr)"
./app/scripts/build-push-image.sh
```