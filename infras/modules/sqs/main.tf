# SQS Queue
resource "aws_sqs_queue" "this" {
  name                       = var.name
  delay_seconds              = var.delay_seconds
  max_message_size           = var.max_message_size
  message_retention_seconds  = var.message_retention_seconds
  receive_wait_time_seconds  = var.receive_wait_time_seconds
  visibility_timeout_seconds = var.visibility_timeout_seconds
  fifo_queue                 = var.fifo_queue
  content_based_deduplication = var.content_based_deduplication
  kms_master_key_id          = var.kms_master_key_id
  kms_data_key_reuse_period_seconds = var.kms_data_key_reuse_period_seconds

  tags = merge(
    var.tags,
    {
      Name = var.name
    }
  )
}

# Dead Letter Queue (optional)
resource "aws_sqs_queue" "dlq" {
  count = var.create_dlq ? 1 : 0

  name                       = "${var.name}-dlq"
  message_retention_seconds  = var.dlq_message_retention_seconds
  kms_master_key_id          = var.kms_master_key_id

  tags = merge(
    var.tags,
    {
      Name = "${var.name}-dlq"
    }
  )
}

# Redrive Policy
resource "aws_sqs_queue_redrive_policy" "this" {
  count = var.create_dlq ? 1 : 0

  queue_url = aws_sqs_queue.this.id
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.dlq[0].arn
    maxReceiveCount     = var.max_receive_count
  })
}

# Queue Policy
resource "aws_sqs_queue_policy" "this" {
  count = var.queue_policy != null ? 1 : 0

  queue_url = aws_sqs_queue.this.id
  policy    = var.queue_policy
}
