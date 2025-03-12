locals {
  target_topic = {
    partitions_count           = coalesce(var.target_topic.partitions_count, 1)
    retention_period_hours     = coalesce(var.target_topic.retention_period_hours, 1)
    partition_write_speed_kbps = coalesce(var.target_topic.partition_write_speed_kbps, 128)
    consumers                  = coalesce(var.target_topic.consumers, {})
  }
}

resource "yandex_datatransfer_endpoint" "cdc_source_endpoint" {
  name = "${var.name}-cdc-source"
  settings {
    ydb_source {
      database               = var.source_db_path
      service_account_id     = var.source_db_reader_sa_id
      paths                  = [var.source_db_table_path]
      changefeed_custom_name = var.changefeed_custom_name
    }
  }
}
resource "yandex_datatransfer_endpoint" "cdc_target_endpoint" {
  name = "${var.name}-cdc-target"
  settings {
    yds_target {
      database           = var.target_db_path
      service_account_id = var.target_db_writer_sa_id
      stream             = yandex_ydb_topic.cdc_target_topic.name
      serializer {
        serializer_auto {}
      }
    }
  }
}

resource "yandex_datatransfer_transfer" "cdc_transfer" {
  name      = var.name
  source_id = yandex_datatransfer_endpoint.cdc_source_endpoint.id
  target_id = yandex_datatransfer_endpoint.cdc_target_endpoint.id
  type      = "INCREMENT_ONLY"
  runtime {
    yc_runtime {
      job_count = 1
      upload_shard_params {
        process_count = 1
        job_count     = 4
      }
    }
  }
}

resource "yandex_ydb_topic" "cdc_target_topic" {
  database_endpoint = var.target_db_endpoint
  name              = "${var.name}-cdc-target"

  supported_codecs       = []
  partitions_count       = local.target_topic.partitions_count
  retention_period_hours = local.target_topic.retention_period_hours

  partition_write_speed_kbps = local.target_topic.partition_write_speed_kbps

  dynamic "consumer" {
    for_each = local.target_topic.consumers
    content {
      name = consumer.value.name
    }
  }

  lifecycle {
    // For Trigger
    ignore_changes = [consumer]
  }
}
