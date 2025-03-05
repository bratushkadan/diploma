resource "yandex_datatransfer_endpoint" "test_cdc_source" {
  name = "${var.name}-ydb-to-ydb-source"
  settings {
    ydb_source {
      database               = var.source_db_path
      service_account_id     = var.source_db_reader_sa_id
      paths                  = var.source_db_table_paths
      changefeed_custom_name = var.changefeed_custom_name
    }
  }
}
resource "yandex_datatransfer_endpoint" "test_cdc_target" {
  name = "${var.name}-ydb-to-ydb-target"
  settings {
    ydb_target {
      database           = var.target_db_path
      service_account_id = var.target_db_writer_sa_id
      # TODO:
      # path = ""
    }
  }
}

resource "yandex_datatransfer_transfer" "test_cdc_transfer" {
  name      = "${var.name}-ydb-to-ydb-transfer"
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
