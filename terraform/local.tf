locals {
  default_zone = "ru-central1-b"
  folder_id    = "b1gtf4tsi5mttbuirmbv"

  common_name = "serverless-ymq-type-standard"

  sqs_queues = {
    account_creations   = "account-creations"
    email_confirmations = "email-confirmations"
  }
}
