resource "yandex_ydb_database_serverless" "this" {
  name        = "${local.common_name}-db"
  description = "ecom services serverless ydb"
}

resource "yandex_ydb_table" "test_cdc_table" {
  path              = "test_cdc_table"
  connection_string = yandex_ydb_database_serverless.this.ydb_full_endpoint

  column {
    name     = "id"
    type     = "Bigserial"
    not_null = true
  }
  column {
    name     = "meta"
    type     = "Json"
    not_null = true
  }
  column {
    name     = "created_at"
    type     = "Timestamp"
    not_null = true
  }
  primary_key = ["id"]
}

// TODO: create a Terraform module for this scenario

// See for Data Transfer endpoints configuration https://yandex.cloud/ru/docs/data-transfer/operations/endpoint/source/ydb#endpoint-settings
resource "yandex_ydb_table_changefeed" "ydb_changefeed" {
  /* table_id here is like:
  grpcs://ydb.serverless.yandexcloud.net:2135/?database=/ru-central1/<cloud_name>/<ydb_database_id>?path=<table_name>
  */
  table_id = yandex_ydb_table.test_cdc_table.id
  name     = "changefeed"
  mode     = "NEW_IMAGE"
  format   = "JSON"

  retention_period = "PT1H"

  consumer {
    // See for Data Transfer endpoints configuration https://yandex.cloud/ru/docs/data-transfer/operations/endpoint/source/ydb#endpoint-settings
    // (this particular setting)
    name = "__data_transfer_consumer"
  }
}

resource "yandex_datatransfer_endpoint" "test_cdc_source" {
  name = "test-cdc-source"
  settings {
    ydb_source {
      database           = yandex_ydb_database_serverless.this.database_path
      service_account_id = yandex_iam_service_account.app.id
      paths = [
        yandex_ydb_table.test_cdc_table.path
      ]
      changefeed_custom_name = yandex_ydb_table_changefeed.ydb_changefeed.name
    }
  }
}
resource "yandex_datatransfer_endpoint" "test_cdc_target" {
  name = "test-cdc-target"
  settings {
    yds_target {
      database           = yandex_ydb_database_serverless.this.database_path
      service_account_id = yandex_iam_service_account.app.id
      stream             = yandex_ydb_topic.test_topic.name
      serializer {
        serializer_auto {}
      }
    }
  }
}

resource "yandex_datatransfer_transfer" "test_cdc_transfer" {
  name      = "test-cdc-transfer"
  source_id = yandex_datatransfer_endpoint.test_cdc_source.id
  target_id = yandex_datatransfer_endpoint.test_cdc_target.id
  type      = "SNAPSHOT_AND_INCREMENT"
  runtime {
    yc_runtime {
      job_count = 1
      upload_shard_params {
        process_count = 1
        job_count     = 1
      }
    }
  }
}

resource "yandex_ydb_topic" "test_topic" {
  database_endpoint = yandex_ydb_database_serverless.this.ydb_full_endpoint
  name              = "test-topic"

  supported_codecs       = ["raw", "gzip"]
  partitions_count       = 1
  retention_period_hours = 1

  partition_write_speed_kbps = 128

  consumer {
    name             = "test-topic-consumer"
    supported_codecs = ["raw", "gzip"]
  }
}
