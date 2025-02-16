data "yandex_lockbox_secret" "app_sa_static_key" {
  secret_id = resource.yandex_lockbox_secret.app_sa_static_key.id
}
data "yandex_lockbox_secret" "email_provider" {
  name = "yandex-mail-provider"
}
data "yandex_lockbox_secret" "token_infra" {
  name = "token-ids-infra"
}


locals {
  versions = {
    auth = {
      account            = "0.0.5"
      email_confirmation = "0.0.5"
    }
  }

  // No way to get around circular dependencies with YMQ Trigger :(
  email_confirmation_origin = "https://d5d0b63n81bf2dbcn9q6.z7jmlavt.apigw.yandexcloud.net"

  containers = {
    auth = {
      account = {
        count = local.versions.auth.account == "" ? 0 : 1
      }
      email_confirmation = {
        count = local.versions.auth.email_confirmation == "" ? 0 : 1
      }
    }
  }

  env = tomap({ for _, v in [
    "YDB_ENDPOINT",
    "YDB_DOC_API_ENDPOINT",
    "AWS_ACCESS_KEY_ID",
    "AWS_SECRET_ACCESS_KEY",
    "SQS_QUEUE_URL_EMAIL_CONFIRMATIONS",
    "SQS_QUEUE_URL_ACCOUNT_CREATIONS",
    "SENDER_EMAIL",
    "SENDER_PASSWORD",
    "EMAIL_CONFIRMATION_API_ENDPOINT",
    "EMAIL_CONFIRMATION_ORIGIN",
    "APP_ID_ACCOUNT_HASH_SALT",
    "APP_ID_TOKEN_HASH_SALT",
    "APP_PASSWORD_HASH_SALT",
    "APP_AUTH_TOKEN_PRIVATE_KEY",
    "APP_AUTH_TOKEN_PUBLIC_KEY",

    "YMQ_TRIGGER_HTTP_ENDPOINTS_ENABLED",

    // LEGACY
    "SQS_ENDPOINT",
  ] : v => v })

  lockbox = {
    auth = {
      account = [
        {
          id                   = data.yandex_lockbox_secret.app_sa_static_key.id
          version_id           = data.yandex_lockbox_secret.app_sa_static_key.current_version[0].id
          key                  = "access_key_id"
          environment_variable = local.env.AWS_ACCESS_KEY_ID
        },
        {
          id                   = data.yandex_lockbox_secret.app_sa_static_key.id
          version_id           = data.yandex_lockbox_secret.app_sa_static_key.current_version[0].id
          key                  = "secret_access_key"
          environment_variable = local.env.AWS_SECRET_ACCESS_KEY
        },
        {
          id                   = data.yandex_lockbox_secret.token_infra.id
          version_id           = data.yandex_lockbox_secret.token_infra.current_version[0].id
          key                  = "auth_account_id_hash_salt"
          environment_variable = local.env.APP_ID_ACCOUNT_HASH_SALT
        },
        {
          id                   = data.yandex_lockbox_secret.token_infra.id
          version_id           = data.yandex_lockbox_secret.token_infra.current_version[0].id
          key                  = "auth_token_id_hash_salt"
          environment_variable = local.env.APP_ID_TOKEN_HASH_SALT
        },
        {
          id                   = data.yandex_lockbox_secret.token_infra.id
          version_id           = data.yandex_lockbox_secret.token_infra.current_version[0].id
          key                  = "auth_password_hash_salt"
          environment_variable = local.env.APP_PASSWORD_HASH_SALT
        },
        {
          id                   = data.yandex_lockbox_secret.token_infra.id
          version_id           = data.yandex_lockbox_secret.token_infra.current_version[0].id
          key                  = "auth_token_private.key"
          environment_variable = local.env.APP_AUTH_TOKEN_PRIVATE_KEY
        },
        {
          id                   = data.yandex_lockbox_secret.token_infra.id
          version_id           = data.yandex_lockbox_secret.token_infra.current_version[0].id
          key                  = "auth_token_public.key"
          environment_variable = local.env.APP_AUTH_TOKEN_PUBLIC_KEY
        },
      ]
      email_confirmation = [
        {
          id                   = data.yandex_lockbox_secret.app_sa_static_key.id
          version_id           = data.yandex_lockbox_secret.app_sa_static_key.current_version[0].id
          key                  = "access_key_id"
          environment_variable = local.env.AWS_ACCESS_KEY_ID
        },
        {
          id                   = data.yandex_lockbox_secret.app_sa_static_key.id
          version_id           = data.yandex_lockbox_secret.app_sa_static_key.current_version[0].id
          key                  = "secret_access_key"
          environment_variable = local.env.AWS_SECRET_ACCESS_KEY
        },
        {
          id                   = data.yandex_lockbox_secret.email_provider.id
          version_id           = data.yandex_lockbox_secret.email_provider.current_version[0].id
          key                  = "email"
          environment_variable = local.env.SENDER_EMAIL
        },
        {
          id                   = data.yandex_lockbox_secret.email_provider.id
          version_id           = data.yandex_lockbox_secret.email_provider.current_version[0].id
          key                  = "password"
          environment_variable = local.env.SENDER_PASSWORD
        },
      ]
    }
  }
}

resource "yandex_serverless_container" "auth_email_confirmation" {
  count = local.containers.auth.email_confirmation.count

  name        = "auth-email-confirmation"
  description = "email confirmation container for service auth"

  cores              = 1
  core_fraction      = 50
  memory             = 128
  execution_timeout  = "10s"
  service_account_id = yandex_iam_service_account.app.id
  runtime {
    type = "http"
  }

  image {
    url = "cr.yandex/${yandex_container_repository.auth_email_confirmation_repository.name}:${local.versions.auth.email_confirmation}"
    environment = {
      (local.env.YMQ_TRIGGER_HTTP_ENDPOINTS_ENABLED) = "1"
      (local.env.YDB_DOC_API_ENDPOINT)               = yandex_ydb_database_serverless.this.document_api_endpoint
      (local.env.SQS_QUEUE_URL_EMAIL_CONFIRMATIONS)  = yandex_message_queue.email_confirmations.id
      (local.env.EMAIL_CONFIRMATION_API_ENDPOINT)    = local.auth_email_confirmation_api_endpoint
      (local.env.EMAIL_CONFIRMATION_ORIGIN)          = local.email_confirmation_origin
    }
  }

  dynamic "secrets" {
    for_each = toset(local.lockbox.auth.email_confirmation)
    content {
      id                   = secrets.value.id
      version_id           = secrets.value.version_id
      key                  = secrets.value.key
      environment_variable = secrets.value.environment_variable
    }
  }

  depends_on = [
    yandex_resourcemanager_folder_iam_member.app_lockbox_payload_viewer,
    yandex_resourcemanager_folder_iam_member.app_images_puller,
  ]

  log_options {
    min_level = "INFO"
  }
  # provision_policy {
  #   min_instances = 1
  # }
}

resource "yandex_serverless_container_iam_binding" "auth_email_confirmation" {
  count        = local.containers.auth.email_confirmation.count
  container_id = yandex_serverless_container.auth_email_confirmation[0].id
  role         = "serverless.containers.invoker"

  members = [
    "serviceAccount:${yandex_iam_service_account.auth_caller.id}",
  ]
}

resource "yandex_function_trigger" "auth_account_creation" {
  count       = local.containers.auth.email_confirmation.count
  name        = "auth-account-creation"
  description = "trigger for calling email confirmation serverless containers when account creation message is published to YMQ"

  container {
    id                 = yandex_serverless_container.auth_email_confirmation[0].id
    service_account_id = yandex_iam_service_account.auth_caller.id
    path               = "/api/v1/auth:send-confirmation-email-trigger"
  }

  message_queue {
    queue_id           = yandex_message_queue.account_creations.arn
    service_account_id = yandex_iam_service_account.app.id
    batch_cutoff       = "0"
    batch_size         = 1
  }
}

resource "yandex_serverless_container" "auth_account" {
  count = local.containers.auth.account.count

  name        = "auth-account"
  description = "account / tokens container for service auth"

  cores              = 1
  core_fraction      = 50
  memory             = 128
  execution_timeout  = "10s"
  service_account_id = yandex_iam_service_account.app.id
  runtime {
    type = "http"
  }

  image {
    url = "cr.yandex/${yandex_container_repository.auth_account_repository.name}:${local.versions.auth.account}"
    environment = {
      (local.env.YDB_ENDPOINT)                    = yandex_ydb_database_serverless.this.ydb_full_endpoint
      (local.env.SQS_QUEUE_URL_ACCOUNT_CREATIONS) = yandex_message_queue.account_creations.id
    }
  }

  dynamic "secrets" {
    for_each = toset(local.lockbox.auth.account)
    content {
      id                   = secrets.value.id
      version_id           = secrets.value.version_id
      key                  = secrets.value.key
      environment_variable = secrets.value.environment_variable
    }
  }

  depends_on = [
    yandex_resourcemanager_folder_iam_member.app_lockbox_payload_viewer,
    yandex_resourcemanager_folder_iam_member.app_images_puller,
  ]

  # log_options {
  #   min_level = "INFO"
  # }
  # provision_policy {
  #   min_instances = 1
  # }
}

resource "yandex_serverless_container_iam_binding" "auth_account" {
  count        = local.containers.auth.account.count
  container_id = yandex_serverless_container.auth_account[0].id
  role         = "serverless.containers.invoker"

  members = [
    "serviceAccount:${yandex_iam_service_account.auth_caller.id}",
  ]
}


resource "yandex_function_trigger" "auth_account_activation" {
  count       = local.containers.auth.account.count
  name        = "account-activation"
  description = "trigger for calling auth account serverless container for account activation when email confirmation message is published to YMQ"

  container {
    id                 = yandex_serverless_container.auth_account[0].id
    service_account_id = yandex_iam_service_account.auth_caller.id
    path               = "/api/v1/users/:activateAccounts"
  }

  message_queue {
    queue_id           = yandex_message_queue.email_confirmations.arn
    service_account_id = yandex_iam_service_account.app.id
    batch_cutoff       = "3"
    batch_size         = 5
  }
}
