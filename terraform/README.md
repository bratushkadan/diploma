# Terraform

> [!TIP]
> To apply Terraform visit "Setup > Terraform" on this page.

## Requirement

Run Terraform from this directory. `shell`-provider scripts depend on this pathing to work correctly.

## SLS Containers code release / update logic

1. Build a docker container.
2. Upload docker container to registry.
3. Bump tag in the Terraform and apply Terraform configuration.

> [!TIP]
> There are of course better options.

## Setup

### The very initial setup

#### Email provider credentials

Create Lockbox secret to send emails with the following fields:
- `email` (sender);
- `password`.

Paste it's id to `data.yandex_lockbox_secret.email_provider` Terraform resource:

```terraform
data "yandex_lockbox_secret" "email_provider" {
  secret_id = ""
}
```

#### JWT public/private keys & id hash salts

Generate JWT signing/verifying keys first:

```sh
openssl ecparam -genkey -name prime256v1 -noout -out private_key.pem
```

```sh
openssl ec -in private_key.pem -pubout -out public_key.pem
```

Create Lockbox secret `token-ids-infra` with the following fields:

- `auth_token_private.key`;
- `auth_token_public.key`;
- `auth_password_hash_salt`;
- `auth_token_id_hash_salt`;
- `auth_account_id_hash_salt`.

Paste it's id to `data.yandex_lockbox_secret.token_infra` Terraform resource:

```terraform
data "yandex_lockbox_secret" "token_infra" {
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


## Infrastructure Roadmap (some details are obsolete)

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

