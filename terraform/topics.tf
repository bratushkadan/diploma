resource "yandex_ydb_topic" "cart_contents_publish_requests" {
  database_endpoint = yandex_ydb_database_serverless.this.ydb_full_endpoint
  name              = "cart/cart_contents_publish_requests_topic"
  description       = "topic for messages requesting a cart contents publish"

  supported_codecs       = []
  partitions_count       = 1
  retention_period_hours = 1

  partition_write_speed_kbps = 128

  # consumer {
  #   name = "foo"
  # }
}
resource "yandex_ydb_topic" "cart_contents" {
  database_endpoint = yandex_ydb_database_serverless.this.ydb_full_endpoint
  name              = "cart/cart_contents_topic"
  description       = "topic for messages of cart contents"

  supported_codecs       = []
  partitions_count       = 1
  retention_period_hours = 1

  partition_write_speed_kbps = 128

  # consumer {
  #   name = "foo"
  # }
}
resource "yandex_ydb_topic" "cart_clear_requests" {
  database_endpoint = yandex_ydb_database_serverless.this.ydb_full_endpoint
  name              = "cart/cart_clear_requests_topic"
  description       = "cart clear requests upon completed order"

  supported_codecs       = []
  partitions_count       = 1
  retention_period_hours = 1

  partition_write_speed_kbps = 128

  # consumer {
  #   name = "foo"
  # }
}

resource "yandex_ydb_topic" "products_reservations" {
  database_endpoint = yandex_ydb_database_serverless.this.ydb_full_endpoint
  name              = "products/products_reservartions_topic"
  description       = "topic for products reservation messages"

  supported_codecs       = []
  partitions_count       = 1
  retention_period_hours = 1

  partition_write_speed_kbps = 128

  # consumer {
  #   name = "foo"
  # }
}
resource "yandex_ydb_topic" "products_reserved" {
  database_endpoint = yandex_ydb_database_serverless.this.ydb_full_endpoint
  name              = "products/reserved_products_topic"
  description       = "topic for reserved products messages"

  supported_codecs       = []
  partitions_count       = 1
  retention_period_hours = 1

  partition_write_speed_kbps = 128
}
resource "yandex_ydb_topic" "products_unreservations" {
  database_endpoint = yandex_ydb_database_serverless.this.ydb_full_endpoint
  name              = "products/products_unreservations_topic"
  description       = "topic for products unreservations messages"

  supported_codecs       = []
  partitions_count       = 1
  retention_period_hours = 1

  partition_write_speed_kbps = 128
}
