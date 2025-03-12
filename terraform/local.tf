locals {
  default_zone = "ru-central1-b"
  folder_id    = "b1gtf4tsi5mttbuirmbv"

  common_name = "ecom"

  opensearch_subnet_cidr                    = "172.16.0.0/24"
  opensearch_instance_data_disk_device_name = "opensearch-data"

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
