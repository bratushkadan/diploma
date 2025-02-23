resource "random_id" "bucket_ecom" {
  byte_length = 8
}

resource "yandex_storage_bucket" "ecom" {
  folder_id = local.folder_id
  bucket    = "ecom-${random_id.bucket_ecom.hex}"

  acl = "public-read"

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET"]
    allowed_origins = ["*"]
    expose_headers  = ["ETag"]
    max_age_seconds = 3000
  }
}
