locals {
  changefeed_flattened = !var.ydb_migrations_applied ? {} : merge([for k, s in local.ydb_tables : { for tk, t in s : "${k}_${tk}" => t if t.cdc.enabled }]...)
}

// See for Data Transfer endpoints configuration https://yandex.cloud/ru/docs/data-transfer/operations/endpoint/source/ydb#endpoint-settings
resource "yandex_ydb_table_changefeed" "ydb_changefeed" {
  for_each = local.changefeed_flattened
  /* table_id here is like:
  grpcs://ydb.serverless.yandexcloud.net:2135/?database=/ru-central1/<cloud_name>/<ydb_database_id>?path=<table_name>
  */
  connection_string = yandex_ydb_database_serverless.this.ydb_full_endpoint
  table_path        = each.value.path
  name              = "changefeed"
  mode              = "NEW_IMAGE"
  format            = "JSON"

  retention_period = "PT1H"

  consumer {
    // See for Data Transfer endpoints configuration https://yandex.cloud/ru/docs/data-transfer/operations/endpoint/source/ydb#endpoint-settings
    // (this particular setting)
    name = "__data_transfer_consumer"
  }
}

module "cdc" {
  for_each = local.changefeed_flattened
  source   = "./modules/cdc"

  name                   = "products"
  source_db_path         = yandex_ydb_database_serverless.this.database_path
  source_db_reader_sa_id = yandex_iam_service_account.app.id
  changefeed_custom_name = yandex_ydb_table_changefeed.ydb_changefeed["products_products"].name
  target_db_path         = yandex_ydb_database_serverless.this.database_path
  target_db_endpoint     = yandex_ydb_database_serverless.this.ydb_full_endpoint
  target_db_writer_sa_id = yandex_iam_service_account.app.id
  source_db_table_path   = local.ydb_tables.products.products.path
  target_topic = {
    consumers = each.value.consumers
  }
}
