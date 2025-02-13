locals {
  versions = {
    auth = {
      accounts           = ""
      email_confirmation = "0.0.1-rc2"
    }
  }
}

locals {
  lockbox = {
    auth = {
      accounts = []
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
  count = local.versions.auth.email_confirmation != "" ? 1 : 0

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
  provision_policy {
    min_instances = 1
  }
}

resource "yandex_function_trigger" "auth_account_creation" {
  count       = local.versions.auth.email_confirmation != "" ? 1 : 0
  name        = "auth-account-creation"
  description = "trigger for calling email confirmation serverless containers when email confirmation message is published to YMQ"

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
