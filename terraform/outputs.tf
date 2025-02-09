output "ydb" {
  value = {
    id                    = yandex_ydb_database_serverless.this.id
    document_api_endpoint = yandex_ydb_database_serverless.this.document_api_endpoint
    full_endpoint         = yandex_ydb_database_serverless.this.ydb_full_endpoint
    api_endpoint          = yandex_ydb_database_serverless.this.ydb_api_endpoint
    database_path         = yandex_ydb_database_serverless.this.database_path
  }
}

output "ymq" {
  value = {
    queues = {
      account_creations = {
        url = yandex_message_queue.account_creations.id
      }
      email_confirmations = {
        url = yandex_message_queue.email_confirmations.id
      }
    }
  }
}

output "ydb_ymq_manager_static_key_lockbox_secret_id" {
  value = yandex_lockbox_secret.ydb_ymq_manager_sa_static_key.id
}

output "app_sa" {
  value = {
    id                           = yandex_iam_service_account.app.id
    key_id                       = yandex_iam_service_account_key.app_sa.id
    auth_key_lockbox_secret_id   = yandex_lockbox_secret.app_sa_static_key.id
    static_key_lockbox_secret_id = yandex_lockbox_secret.app_sa_static_key.id
  }
}
