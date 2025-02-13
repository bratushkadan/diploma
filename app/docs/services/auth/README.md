# Auth service

## Roadmap

### Application

- [x] ⏳ Add CreateSeller/CreateAdmin service methods
- [x] Create Notification Secondary Adapter
  - [x] Account Creation
  - [x] Email Confirmation
- [ ] Create Notification Primary Adapter
  - [x] Long Polling
    - [x] Account Creation
    - [x] Email Confirmation
  - [ ] Cloud Function (HTTP handler)
    - [ ] Account Creation
    - [ ] Email Confirmation
- [x] Test Notification Secondary Adapter
  - [x] Account Creation
  - [x] Email Confirmation
- [ ] Test Notification Primary Adapter
  - [x] Long Polling
    - [x] Account Creation
    - [x] Email Confirmation
  - [ ] Cloud Function (HTTP handler)
    - [ ] Account Creation
    - [ ] Email Confirmation
- [x] Email Confirmation Service refactoring
- [ ] Wire service and HTTP primary adapter
- [ ] Create Cloud Functions Code Boilerplate
- [ ] Create Cloud Functions Terraform Configuration Code
- [x] ⏳ Write service/infrastructure Integration Tests

### Tests

- [ ] Integration Tests
  - [x] Service
    - [x] `CreateUser`
    - [x] `ActivateAccounts`
    - [x] `Authenticate`
    - [x] `ReplaceRefreshToken`
    - [x] `CreateAccessToken`
    - [x] `CreateSeller`
    - [x] `CreateAdmin`
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
    - [x] Account Created Notifications
      - [x] `Send`
    - [x] Email Confirmed Notifications
      - [x] `Send`
    - [x] Token Provider (JWT)
      - [x] `EncodeRefresh`
      - [x] `DecodeRefresh`
      - [x] `EncodeAccess`
      - [x] `DecodeAccess`
 - [ ] Application-level e2e-tests

## Run

### Common steps

```sh
cd app
```

```sh
TF_OUTPUT=$(../terraform/tf output -json -no-color)
export YDB_SERVICE_ACCOUNT_KEY_FILE_CREDENTIALS="$(scripts/ydb_access_token.sh)"
export YDB_ENDPOINT="$(echo "${TF_OUTPUT}" | jq -cMr .ydb.value.full_endpoint)"
export YDB_DOC_API_ENDPOINT="$(echo "${TF_OUTPUT}" | jq -cMr .ydb.value.document_api_endpoint)"
export YDB_AUTH_METHOD=environ
APP_SA_STATIC_KEY_SECRET_ID="$(echo $TF_OUTPUT | jq -cMr .app_sa.value.static_key_lockbox_secret_id)"
SECRET=$(yc lockbox payload get "${APP_SA_STATIC_KEY_SECRET_ID}")
export AWS_ACCESS_KEY_ID=$(echo $SECRET | yq -M '.entries.[] | select(.key == "access_key_id").text_value')
export AWS_SECRET_ACCESS_KEY=$(echo $SECRET | yq -M '.entries.[] | select(.key == "secret_access_key").text_value')
export SQS_QUEUE_URL_EMAIL_CONFIRMATIONS="$(echo "${TF_OUTPUT}" | jq -cMr .ymq.value.queues.email_confirmations.url)"
export SQS_QUEUE_URL_ACCOUNT_CREATIONS="$(echo "${TF_OUTPUT}" | jq -cMr .ymq.value.queues.account_creations.url)"
source .env
```

### Run integration tests

```sh
go run cmd/auth/integration_tests/main.go
```

### Account confirmation test pipeline example

1\. Run email confirmation consumer in window 1:
```sh
go run cmd/auth/email-confirmation-consumer/main.go
```

2\. Run account creation consumer in window 2:
```sh
go run cmd/auth/account-creation-consumer/main.go
```

3\. Run account creation consumer in window 3:
```sh
TARGET_EMAIL= go run cmd/auth/account-creation-producer/main.go
```

4\. Run confirm email function in window 3
```sh
CONFIRMATION_TOKEN= go run cmd/auth/confirm-email/main.go
```

5\. Run YQL query to cleanup created account

```sql
DELETE FROM accounts
WHERE email = "<TARGET_EMAIL>"
RETURNING *;
```

### Service Pieces

#### Run account creation consumer

```sh
go run cmd/auth/account-creation-consumer/main.go
```

#### Run account creation producer

```sh
TARGET_EMAIL= go run cmd/auth/account-creation-producer/main.go
```

#### Run email confirmation consumer

```sh
go run cmd/auth/email-confirmation-consumer/main.go
```

#### Run confirm email function

```sh
CONFIRMATION_TOKEN= go run cmd/auth/confirm-email/main.go
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
