variable "ydb_additionals" {
  type        = bool
  description = "apply ydb additional settings once YDB database has been created during the previous Terraform apply and all the migrations and table creations have been applied"
  nullable    = false
}
