locals {
  auth_email_confirmation_api_endpoint = "/auth:confirm-email"

  api_gateway = {
    spec_options = {
      auth_email_confirmation_api_endpoint = local.auth_email_confirmation_api_endpoint

      containers = {
        auth = {
          account = {
            id    = local.containers.auth.account.count > 0 ? yandex_serverless_container.auth_account[0].id : ""
            sa_id = yandex_iam_service_account.auth_caller.id
          }
          email_confirmation = {
            id    = local.containers.auth.email_confirmation.count > 0 ? yandex_serverless_container.auth_email_confirmation[0].id : ""
            sa_id = yandex_iam_service_account.auth_caller.id
          }
        }
      }

      functions = {
        send_confirmation_email = {
          function_id = yandex_function.send_confirmation_email.id
          version_tag = local.functions.send_confirmation_email.version
          sa_id       = yandex_iam_service_account.auth_caller.id
        }
        confirm_email = {
          function_id = yandex_function.confirm_email.id
          version_tag = local.functions.confirm_email.version
          sa_id       = yandex_iam_service_account.auth_caller.id
        }
        test_ydb = {
          function_id = yandex_function.test_ydb.id
          version_tag = local.functions.test_ydb.version
          sa_id       = yandex_iam_service_account.auth_caller.id
        }
      }
    }
  }
}

resource "yandex_api_gateway" "auth" {
  name        = "auth-service-api-gw"
  description = "auth service API Gateway"

  spec = templatefile("${path.module}/config/api-gateway-spec.yaml", local.api_gateway.spec_options)
}
