# Products microservice documentation

## Data Model

[DB Diagram](https://dbdiagram.io/d/ecom-67b96d09263d6cf9a01083b2)

YDB Schema:

```sql
CREATE TABLE `products/products` (
    id Uuid NOT NULL,
    seller_id Utf8 NOT NULL,
    name Utf8 NOT NULL,
    description Utf8 NOT NULL,
    pictures Json NOT NULL,
    metadata Json NOT NULL,
    stock Uint32 NOT NULL,
    created_at Datetime NOT NULL,
    updated_at Datetime NOT NULL,
    deleted_at Datetime,
    PRIMARY KEY (id),
    INDEX idx_seller_id GLOBAL ASYNC ON (seller_id),
);
```

## SEED(s) use cases

- Add/Get/List/Update/Delete for Product Entity
  - Filtering by *sellerId* is an important requirement for the List handler;
  - Point lookups using *productId*.
- Upload/Delete image for Product (limit is 2MiB due to serverless containers limitation of 3MiB request size, including http headers)

## Run

### Setup env
```sh
TF_OUTPUT=$(../terraform/tf output -json -no-color)
export YDB_ENDPOINT="$(echo "${TF_OUTPUT}" | jq -cMr .ydb.value.full_endpoint)"
export YDB_SERVICE_ACCOUNT_KEY_FILE_CREDENTIALS="$(scripts/ydb_access_token.sh)"
export YDB_AUTH_METHOD=environ
APP_SA_STATIC_KEY_SECRET_ID="$(echo $TF_OUTPUT | jq -cMr .app_sa.value.static_key_lockbox_secret_id)"
SECRET=$(yc lockbox payload get "${APP_SA_STATIC_KEY_SECRET_ID}")
export AWS_ACCESS_KEY_ID="$(echo $SECRET | yq -M '.entries.[] | select(.key == "access_key_id").text_value')"
export AWS_SECRET_ACCESS_KEY="$(echo $SECRET | yq -M '.entries.[] | select(.key == "secret_access_key").text_value')"
export PICTURES_BUCKET="$(echo $TF_OUTPUT | jq -cMr .s3.value)"
INFRA_TOKENS_SECRET_ID="$(echo $TF_OUTPUT | jq -cMr .infra_tokens_lockbox_secret_id.value)"
INFRA_TOKENS_SECRET="$(yc lockbox payload get "${INFRA_TOKENS_SECRET_ID}")"
export APP_AUTH_TOKEN_PUBLIC_KEY="$(echo $INFRA_TOKENS_SECRET | yq -M '.entries.[] | select(.key == "auth_token_public.key").text_value')"
go run cmd/products/main.go
```

