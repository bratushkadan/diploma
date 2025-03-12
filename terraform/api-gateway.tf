locals {
  auth_email_confirmation_api_endpoint = "/api/v1/auth:confirm-email"

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
        products = {
          id = ""
          # id    = local.containers.products.count > 0 ? yandex_serverless_container.products[0].id : ""
          sa_id = yandex_iam_service_account.auth_caller.id
        }
        catalog = {
          id    = local.containers.catalog.count > 0 ? yandex_serverless_container.catalog[0].id : ""
          sa_id = yandex_iam_service_account.auth_caller.id
        }
        cart = {
          id = ""
          # id    = local.containers.cart.count > 0 ? yandex_serverless_container.cart[0].id : ""
          sa_id = yandex_iam_service_account.auth_caller.id
        }
        orders = {
          id = ""
          # id    = local.containers.orders.count > 0 ? yandex_serverless_container.orders[0].id : ""
          sa_id = yandex_iam_service_account.auth_caller.id
        }
        feedback = {
          id = ""
          # id    = local.containers.feedback.count > 0 ? yandex_serverless_container.feedback[0].id : ""
          sa_id = yandex_iam_service_account.auth_caller.id
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
