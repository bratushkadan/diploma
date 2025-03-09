variable "name" {
  type        = string
  description = "Name of datatransfer and its constituent resources"
  nullable    = false
}

variable "source_db_path" {
  type     = string
  nullable = false
}

variable "source_db_reader_sa_id" {
  type     = string
  nullable = false
}

variable "source_db_table_path" {
  type     = string
  nullable = false
}

variable "changefeed_custom_name" {
  type = string
}

variable "target_db_path" {
  type     = string
  nullable = false
}
variable "target_db_endpoint" {
  type     = string
  nullable = false
}
variable "target_topic" {
  type = object({
    partitions_count           = optional(number)
    retention_period_hours     = optional(number)
    partition_write_speed_kbps = optional(number)
    consumers                  = map(object({ name = string }))
  })
  nullable = false
}
variable "target_db_writer_sa_id" {
  type     = string
  nullable = false
}
