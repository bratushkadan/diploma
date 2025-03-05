resource "yandex_ydb_database_serverless" "this" {
  name        = "${local.common_name}-db"
  description = "ecom services serverless ydb"
}
