locals {
  sqs_queues = {
    account_creations   = "account-creations"
    email_confirmations = "email-confirmations"
  }
}

resource "yandex_message_queue" "email_confirmations" {
  name = local.sqs_queues.email_confirmations

  visibility_timeout_seconds = 30

  redrive_policy = jsonencode({
    deadLetterTargetArn = yandex_message_queue.email_confirmations_dmq.arn
    maxReceiveCount     = 5
  })
}
resource "yandex_message_queue" "email_confirmations_dmq" {
  name = "${local.sqs_queues.email_confirmations}-dmq"
}

resource "yandex_message_queue" "account_creations" {
  name = local.sqs_queues.account_creations

  visibility_timeout_seconds = 30

  redrive_policy = jsonencode({
    deadLetterTargetArn = yandex_message_queue.account_creations_dmq.arn
    maxReceiveCount     = 5
  })
}
resource "yandex_message_queue" "account_creations_dmq" {
  name = "${local.sqs_queues.account_creations}-dmq"
}
