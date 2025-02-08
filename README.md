# Auth service for e-com platform

## Roadmap

- [x] Setup YMQ
- [x] Implement Email Confirmation Sender
- [x] Add YMQ publishing to the Email Confirmation Sender
- [x] Implement Email Confirmer
- [x] Implement Confirm User Account (possibly a mock service that just reads the value from the YMQ)
- [x] Setup API Gateway
- [x] Deploy & test the integrations
- [x] Move setup to Terraform
- [x] Implement Advanced Terraform Code Packing
- [x] Learn how IAM authorization works for Cloud Function & Serverless ecosystem
- [x] Deploy Email Confirmation Cloud function
- [x] Add Browser Email Confirmation Endpoit
- [x] Learn if I can get API Gateway Host header via confirmation link sender function in order to break circular dependency "API Gateway (endpoint) <--> Cloud Function (id)"
- [x] Add Rate Limiting to API Gateway
- [x] Add validation to API Gateway

### Auth

- [ ] ⏳ Add CreateSeller/CreateAdmin service methods (or make CreateAccount method polymorphic?)
- [ ] ⏳ Test YMQ/SQS AccountActivation Secondary Adapter
- [ ] ⏳ Wire service and HTTP primary adapter
- [ ] Create Cloud Functions Code Boilerplate
- [ ] Create Cloud Functions Terraform Configuration Code
- [ ] ⏳ Write Integration Tests

#### Tests Roadmap

- [ ] Integration Tests
  - [ ] Service
    - [x] `CreateAccount`
    - [x] `ActivateAccounts`
    - [x] `Authenticate`
    - [x] `ReplaceRefreshToken`
    - [x] `CreateAccessToken`
    - [ ] `CreateSeller`
    - [ ] `CreateAdmin`
  - [x] Secondary Adapters
    - [x] Account YDB
      - [x] `CreateAccount`
      - [x] `FindAccount`
      - [x] `FindAccountByEmail`
      - [x] `CheckAccountCredentials`
      - [x] `ActivateAccountsByEmail`
    - [x] RefreshToken YDB
      - [x] `List`
      - [x] `Add`
      - [x] `Replace`
      - [x] `Delete`
      - [x] `DeleteByAccountId`
    - [ ] AccountActivation YMQ
      - [ ] `Send`
    - [x] Token Provider (JWT)
      - [x] `EncodeRefresh`
      - [x] `DecodeRefresh`
      - [x] `EncodeAccess`
      - [x] `DecodeAccess`


# New workflow with Terraform + yc/aws CLI

## Setup

### The very initial setup

Create Lockbox secret to send emails with the following fields:
- `email` (sender);
- `password`.

Paste it's id to `data.yandex_lockbox_secret.email_provider` Terraform resource:

```terraform
data "yandex_lockbox_secret" "email_provider" {
  secret_id = ""
}
```

### Terraform

1\. First step is to generate Yandex Message Queue (SQS) aws keys:
```sh
./tf apply -target yandex_iam_service_account_static_access_key.ydb_ymq_manager_sa
```

2\. Export the ymq secrets that are required by the [yandex provider](https://terraform-provider.yandexcloud.net/index.html#optional) in order to work with the SQS API using Terraform:
```sh
export MANAGER_LOCKBOX_SEC_ID=$(./tf output -json -no-color | jq -cMr .ydb_ymq_manager_static_key_lockbox_secret_id.value)
export MANAGER_SECRET=$(yc lockbox payload get "${MANAGER_LOCKBOX_SEC_ID}")
export YC_MESSAGE_QUEUE_ACCESS_KEY=$(echo "${MANAGER_SECRET}" \
  | yq '.entries | .[] | select(.key == "access_key_id").text_value')
export YC_MESSAGE_QUEUE_SECRET_KEY=$(echo "${MANAGER_SECRET}" \
  | yq '.entries | .[] | select(.key == "secret_access_key").text_value')
```

4\. You may run `./tf` without `-target`'ing now.

## Setup credentials for application

```sh
TF_OUTPUT=$(./terraform/tf output -json -no-color)
export YDB_DOC_API_ENDPOINT=$(echo $TF_OUTPUT | jq -cMr .ydb.value.document_api_endpoint)
export SQS_ENDPOINT=$(echo $TF_OUTPUT | jq -cMr .ymq.value.queues.email_confirmation.url)
APP_SA_STATIC_KEY_SECRET_ID="$(echo $TF_OUTPUT | jq -cMr .app_sa.value.static_key_lockbox_secret_id)"
SECRET=$(yc lockbox payload get "${APP_SA_STATIC_KEY_SECRET_ID}")
export AWS_ACCESS_KEY_ID=$(echo $SECRET | yq -M '.entries.[] | select(.key == "access_key_id").text_value')
export AWS_SECRET_ACCESS_KEY=$(echo $SECRET | yq -M '.entries.[] | select(.key == "secret_access_key").text_value')
export AWS_DEFAULT_REGION=ru-central1
EMAIL_SECRET=$(yc lockbox payload get yandex-mail-provider)
export SENDER_EMAIL="$(echo $EMAIL_SECRET | yq -M '.entries.[] | select(.key == "email").text_value')"
export SENDER_PASSWORD="$(echo $EMAIL_SECRET | yq -M '.entries.[] | select(.key == "password").text_value')"

export EMAIL_CONFIRMATION_URL="foo.bar"
```

## Test app

1\. Send confirmation email:
```sh
go run cmd/email-confirmation-sender/main.go
```

Copy the confirmation token that was sent to the email.
Change the `./cmd/email-confirmation/email_confirmation.go`'s value with the token.

2\. Run email confirmation:
```sh
go run cmd/email-confirmation/main.go
```

3\. Read messages about the confirmation from queue (mock service that adds "confirmed" records to user accounts by reading messages from confirmation component):
```sh
go run ./cmd/q-reader/main.go
```

## Bootstrap YDB

### Create `email_confirmation_tokens` database

In DynamoDB, you can define a composite primary key by specifying both a partition key (`HASH` key) and a sort key (`RANGE` key). This allows you to uniquely identify items in the table using a combination of these two attributes, which also facilitates more complex query patterns by allowing operations on sets of items with the same partition key.

```bash
export TABLE_CONF_TOKENS_NAME=email_confirmation_tokens
aws dynamodb create-table \
  --table-name "${TABLE_CONF_TOKENS_NAME}" \
  --attribute-definitions \
    AttributeName=token,AttributeType=S \
    AttributeName=email,AttributeType=S \
  --key-schema \
    AttributeName=email,KeyType=HASH \
    AttributeName=token,KeyType=RANGE \
  --global-secondary-indexes "[
    {
      \"IndexName\": \"TokenIndex\",
      \"KeySchema\": [
        {\"AttributeName\": \"token\",\"KeyType\": \"HASH\"}
      ],
      \"Projection\": {
        \"ProjectionType\": \"ALL\"
      },
      \"ProvisionedThroughput\": {
        \"ReadCapacityUnits\": 5,
        \"WriteCapacityUnits\": 5
      }
    }
  ]" \
  --endpoint "$YDB_DOC_API_ENDPOINT"
aws dynamodb update-time-to-live \
    --table-name "${TABLE_CONF_TOKENS_NAME}"  \
    --time-to-live-specification "Enabled=true, AttributeName=expires_at" \
  --endpoint "$YDB_DOC_API_ENDPOINT"
```

# Floral - old docs

## Run

### Auth service

```bash
go run github.com/bratushkadan/floral/cmd/auth
```

#### Or this way

```bash
AUTH_JWT_PRIVATE_KEY_PATH=./pkg/auth/test_fixtures/private.key AUTH_JWT_PUBLIC_KEY_PATH=./pkg/auth/test_fixtures/public.key YANDEX_MAIL_APP_PASSWORD=<password> go run ./cmd/auth/main.go
```

## Email Sending

1. https://id.yandex.ru/security/app-passwords - add password

[Official Yandex Mail docs](https://yandex.ru/support/yandex-360/business/mail/ru/web/security/oauth)

## HTTP API Docs

### Auth

#### Error Response

Sample:
```json
{
  "errors": [
    {
      "code": 1,
      "message": "bad request"
    }
  ]
}
```

#### Endpoints

##### `POST /api/v1/users/:register`

Request sample:
```json
{
  "name": "danila",
  "email": "foobar@yahoo.com",
  "password": "secretpass123"
}
```

Response sample:
```json
{
  "id": "i1qwk6jcuwjeqzy",
  "name": "danila"
}
```

##### `POST /api/v1/users/:registerSeller`

**admin only**

Request sample:
```json
{
  "seller": {
    "name": "danila",
    "email": "foobar@yahoo.com",
    "password": "secretpass123"
  },
  "access_token": "..."
}
```

Response sample:
```json
{
  "id": "i1qwk6jcuwjeqzy",
  "name": "danila"
}
```

##### `POST /api/v1/users/:registerAdmin`

**admin only** - expose this endpoint with **extreme caution**

Request sample:
```json
{
  "admin": {
    "name": "danila",
    "email": "foobar@yahoo.com",
    "password": "secretpass123"
  },
  "secret_token": "..."
}
```

Response sample:
```json
{
  "id": "i1qwk6jcuwjeqzy",
  "name": "danila"
}
```

##### `POST /api/v1/users/:authenticate`

Request sample:
```json
{
  "email": "foobar@yahoo.com",
  "password": "secretpass123"
}
```

Response sample:
```json
{
  "refresh_token": "..."
}
```

##### `POST /api/v1/users/:renewRefreshToken`

Request sample:
```json
{
  "refresh_token": "..."
}
```

Response sample:
```json
{
  "refresh_token": "..."
}
```

##### `POST /api/v1/users/:createAccessToken`

Request sample:
```json
{
  "refresh_token": "..."
}
```

Response sample:
```json
{
  "access_token": "...",
  "expires_at": "2025-01-06T21:02:13+03:00"
}
```
