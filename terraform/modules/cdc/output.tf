output "target_topic" {
  value = {
    database_endpoint = yandex_ydb_topic.cdc_target_topic.database_endpoint
    name              = yandex_ydb_topic.cdc_target_topic.name
  }
}
output "topic" {
  value = yandex_ydb_topic.cdc_target_topic
}
