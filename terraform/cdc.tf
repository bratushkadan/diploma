// See for Data Transfer endpoints configuration https://yandex.cloud/ru/docs/data-transfer/operations/endpoint/source/ydb#endpoint-settings
resource "yandex_ydb_table_changefeed" "ydb_changefeed" {
  for_each = !var.ydb_additionals ? {} : merge([for k, s in local.ydb_tables : { for tk, t in s : "${k}_${tk}" => t if t.cdc.enabled }]...)

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
  dynamic "consumer" {
    for_each = each.value.consumers
    content {
      name = consumer.value.name
    }
  }
}

# module "cdc_transfer" {
#   source = "./modules/datatransfer-ydb-to-ydb"
# 
#   name                   = "first-test"
#   source_db_path         = yandex_ydb_database_serverless.this.database_path
#   source_db_reader_sa_id = yandex_iam_service_account.app.id
#   changefeed_custom_name = yandex_ydb_table_changefeed.ydb_changefeed["auth_accounts"].name
#   target_db_path         = yandex_ydb_database_serverless.this.database_path
#   target_db_writer_sa_id = yandex_iam_service_account.app.id
#   source_db_table_paths  = [local.ydb_tables.auth.accounts.path]
# 
#   depends_on = [yandex_resourcemanager_folder_iam_member.app_datatransfer_ydb_to_ydb]
# }

### resource "yandex_datatransfer_endpoint" "test_cdc_source" {
###   name  = "test-cdc-source"
###   settings {
###     ydb_source {
###       database           = yandex_ydb_database_serverless.this.database_path
###       service_account_id = yandex_iam_service_account.app.id
###       paths = []
###       changefeed_custom_name = yandex_ydb_table_changefeed.ydb_changefeed.name
###     }
###   }
### }
### resource "yandex_datatransfer_endpoint" "test_cdc_target" {
###   name  = "test-cdc-target"
###   settings {
###     yds_target {
###       database           = yandex_ydb_database_serverless.this.database_path
###       service_account_id = yandex_iam_service_account.app.id
###       stream             = yandex_ydb_topic.test_topic.name
###       serializer {
###         serializer_auto {}
###       }
###     }
###   }
### }
### 
### resource "yandex_datatransfer_transfer" "test_cdc_transfer" {
###   name      = "test-cdc-transfer"
###   source_id = yandex_datatransfer_endpoint.test_cdc_source.id
###   target_id = yandex_datatransfer_endpoint.test_cdc_target.id
###   type      = "SNAPSHOT_AND_INCREMENT"
###   runtime {
###     yc_runtime {
###       job_count = 1
###       upload_shard_params {
###         process_count = 1
###         job_count     = 1
###       }
###     }
###   }
### }


# resource "yandex_ydb_topic" "test_topic" {
#   count             = var.ydb_additionals ? 1 : 0
#   database_endpoint = yandex_ydb_database_serverless.this.ydb_full_endpoint
#   name              = "test-topic"
# 
#   supported_codecs       = ["raw", "gzip"]
#   partitions_count       = 1
#   retention_period_hours = 1
# 
#   partition_write_speed_kbps = 128
# 
#   consumer {
#     name             = "test-topic-consumer"
#     supported_codecs = ["raw", "gzip"]
#   }
# }
