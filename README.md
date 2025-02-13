# E-com platform

## Services

### [Auth](./app/docs/services/auth)

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

3\. You may run `./tf` without `-target`'ing now.

### Per-service setup

Don't forget to setup YDB:
- [document API for Auth](./app/docs/services/auth)
- [regular for Auth](./app/Makefile)

## Email Sending

1. https://id.yandex.ru/security/app-passwords - add password

[Official Yandex Mail docs](https://yandex.ru/support/yandex-360/business/mail/ru/web/security/oauth)
