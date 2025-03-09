locals {
  default_zone = "ru-central1-b"
  folder_id    = "b1gtf4tsi5mttbuirmbv"

  common_name = "ecom"

  ydb_tables = {
    # auth = {
    #   accounts = {
    #     path = "auth/accounts"
    #     cdc = {
    #       enabled = true
    #     }
    #     consumers = [
    #       {
    #         name = "catalog"
    #       },
    #     ]
    #   }
    # }
    products = {
      products = {
        path = "products/products"
        cdc = {
          enabled = true
        }
        consumers = {
          "catalog" = {
            name = "catalog"
          }
        }
      }
    }
  }
  ydb_doc_tables = {
    auth = {
      tokens = "auth/tokens"
    }
  }
}
