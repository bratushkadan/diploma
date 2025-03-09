# Products microservice documentation

## Data Model

[DB Diagram](https://dbdiagram.io/d/ecom-67b96d09263d6cf9a01083b2)

YDB Schema:

```sql
CREATE TABLE `products/products` (
    id String NOT NULL,
    seller_id Utf8 NOT NULL,
    name Utf8 NOT NULL,
    description Utf8 NOT NULL,
    pictures Json NOT NULL,
    metadata Json NOT NULL,
    stock Uint32 NOT NULL,
    price Double NOT NULL,
    created_at Datetime NOT NULL,
    updated_at Datetime NOT NULL,
    deleted_at Datetime,
    PRIMARY KEY (id),
    INDEX idx_seller_id GLOBAL ASYNC ON (seller_id),
    INDEX idx_created_at_id GLOBAL ASYNC ON (created_at, id)
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
## CURLs for testing

### Get access token

```sh
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"email": "<Email>", "password": "<Password>"}' \
  https://d5d0b63n81bf2dbcn9q6.z7jmlavt.apigw.yandexcloud.net/api/v1/users/:authenticate \
```

```sh
curl -s -X POST \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": ""}' \
  https://d5d0b63n81bf2dbcn9q6.z7jmlavt.apigw.yandexcloud.net/api/v1/users/:createAccessToken \
  | jq -cMr .access_token
```

Or export it:

```sh
ACCESS_TOKEN="$(curl -s -X POST \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": ""}' \
  https://d5d0b63n81bf2dbcn9q6.z7jmlavt.apigw.yandexcloud.net/api/v1/users/:createAccessToken \
  | jq -cMr .access_token)"
```

### Products

#### Create

Sample request:

```sh
curl -sL \
  -X POST \
  -H "Content-Type: application/json" \
  -H "X-Authorization: Bearer ${ACCESS_TOKEN}" \
  -d '{"name": "foo", "stock": 5, "price": 145, "metadata": {}, "description": ""}' http://localhost:8080/api/v1/products | jq
```

Sample response:

```json
{
  "created_at": "2025-02-23T06:28:46+03:00",
  "description": "",
  "id": "13eb5b42-795d-4b18-886d-84b985180c34",
  "metadata": {},
  "name": "foo",
  "pictures": [],
  "price": 145,
  "seller_id": "12dl52q59z8r",
  "stock": 5,
  "updated_at": "2025-02-23T06:28:46+03:00"
}
```

Sample response (insufficient permissions):

```json
{
  "errors": [
    {
      "code": 124,
      "message": "permission denied"
    }
  ]
}
```

#### Get

Sample request:

```sh
curl -sL \
  http://localhost:8080/api/v1/products/31adfeee-574d-4771-bf4c-b6fab6013853 | jq
```

Sample response:

```json
{
  "created_at": "2025-02-23T18:53:49+03:00",
  "description": "",
  "id": "31adfeee-574d-4771-bf4c-b6fab6013853",
  "metadata": {},
  "name": "foo",
  "pictures": [
    {
      "id": "7bbaf374-6566-479e-a547-a1ac63d2e151",
      "url": "https://storage.yandexcloud.net/ecom-57a07237dfa8db13/product-pictures/31adfeee-574d-4771-bf4c-b6fab6013853/7bbaf374-6566-479e-a547-a1ac63d2e151.jpg"
    },
    {
      "id": "d25e68b9-b64e-4930-82fb-fd05986e57da",
      "url": "https://storage.yandexcloud.net/ecom-57a07237dfa8db13/product-pictures/31adfeee-574d-4771-bf4c-b6fab6013853/d25e68b9-b64e-4930-82fb-fd05986e57da.jpg"
    }
  ],
  "seller_id": "12dl52q59z8r",
  "stock": 5,
  "updated_at": "2025-02-23T19:02:39+03:00"
}
```

#### Update

Sample request:

```sh
curl -s -X PATCH \
  -H "X-Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"stock_delta": -2, "metadata":{"foo":"bar", "brand":2025}}' \
  http://localhost:8080/api/v1/products/31adfeee-574d-4771-bf4c-b6fab6013853 | jq
```

Sample response:

```json
{
  "metadata": {
    "brand": 2025,
    "foo": "bar"
  },
  "stock": 1
}
```

Sample error response:

```json
{
  "errors": [
    {
      "code": 0,
      "message": "stock is not sufficient: trying to withdraw 2 units from stock when there's only 1"
    }
  ]
}
```

#### Delete

Sample request:

```sh
curl -s -X DELETE \
  -H "X-Authorization: Bearer ${ACCESS_TOKEN}" \
  http://localhost:8080/api/v1/products/31adfeee-574d-4771-bf4c-b6fab6013853 | jq
```

Sample response:

```json
{
  "id": "31adfeee-574d-4771-bf4c-b6fab6013853"
}
```

Sample error response:

```json
{
  "errors": [
    {
      "code": 0,
      "message": "stock is not sufficient: trying to withdraw 2 units from stock when there's only 1"
    }
  ]
}
```

### Product Images

#### Add

Sample request:

```sh
curl -s -X POST \
  -H "Content-Type: multipart/form-data" \
  -H "X-Authorization: Bearer ${ACCESS_TOKEN}" \
  -F "file=@5-1.jpg" \
  -F "caption=Sample Movie file" \
  http://localhost:8080/api/v1/products/31adfeee-574d-4771-bf4c-b6fab6013853/pictures | jq
```

Sample response:

```json
{
  "id": "7bbaf374-6566-479e-a547-a1ac63d2e151",
  "url": "https://storage.yandexcloud.net/ecom-57a07237dfa8db13/product-pictures/31adfeee-574d-4771-bf4c-b6fab6013853/7bbaf374-6566-479e-a547-a1ac63d2e151.jpg"
}
```

#### Delete

Sample request:

```sh
curl -s -X DELETE \
  -H "X-Authorization: Bearer ${ACCESS_TOKEN}" \
  http://localhost:8080/api/v1/products/6660d375-586c-442d-a42b-0198bea36d3b/pictures/895a5c01-2d44-47ca-b9a1-f3bf8ad4560e | jq
```

Sample response:

```json
{"id":"2c345e94-6409-462c-a412-bf96ae7d9f89"}
```
