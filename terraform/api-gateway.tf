locals {
  auth_email_confirmation_api_endpoint = "/api/v1/auth/confirmEmail"

  api_gateway = {
    spec_options = {
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
          id    = local.containers.products.count > 0 ? yandex_serverless_container.products[0].id : ""
          sa_id = yandex_iam_service_account.auth_caller.id
        }
        catalog = {
          id    = local.containers.catalog.count > 0 ? yandex_serverless_container.catalog[0].id : ""
          sa_id = yandex_iam_service_account.auth_caller.id
        }
        cart = {
          id    = local.containers.cart.count > 0 ? yandex_serverless_container.cart[0].id : ""
          sa_id = yandex_iam_service_account.auth_caller.id
        }
        orders = {
          id    = local.containers.orders.count > 0 ? yandex_serverless_container.orders[0].id : ""
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

  _api_gateway_spec                 = yamldecode(templatefile("${path.module}/config/api-gateway-spec.yaml", local.api_gateway.spec_options))
  api_gateway_spec_with_private_api = yamlencode(local._api_gateway_spec)
  api_gateway_spec = yamlencode(
    merge(local._api_gateway_spec, {
      paths = { for k, v in local._api_gateway_spec.paths : k => v if try(v.x-private-api, false) == false },
      components = merge(local._api_gateway_spec.components, {
        schemas   = { for k, v in local._api_gateway_spec.components.schemas : k => v if contains(try(v.x-tags, []), "private_api") == false },
        responses = { for k, v in local._api_gateway_spec.components.responses : k => v if contains(try(v.x-tags, []), "private_api") == false },
      })
    })
  )
}

resource "local_file" "public_private_api_oapi_spec" {
  filename = "${path.module}/../app/public_private_api_oapi_spec.yaml"
  content  = local.api_gateway_spec_with_private_api
}

resource "yandex_api_gateway" "auth" {
  name        = "auth-service-api-gw"
  description = "auth service API Gateway"

  spec = local.api_gateway_spec
}
