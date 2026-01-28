# SQS Queue
resource "aws_sqs_queue" "main" {
  name                       = "${var.project_name}-queue"
  delay_seconds              = 0
  max_message_size           = 262144
  message_retention_seconds  = 345600 # 4 days
  receive_wait_time_seconds  = 0
  visibility_timeout_seconds = 360 # 6 minutes (6x the Lambda timeout of 60s)

  tags = {
    Name        = var.project_name
    Environment = "Production"
  }
}

# SQS Dead Letter Queue
resource "aws_sqs_queue" "dead_letter" {
  name                      = "${var.project_name}-dlq"
  delay_seconds             = 0
  max_message_size          = 262144
  message_retention_seconds = 1209600 # 14 days
  receive_wait_time_seconds = 0

  tags = {
    Name        = var.project_name
    Environment = "Production"
  }
}

# Redrive Policy - Send failed messages to DLQ
resource "aws_sqs_queue_redrive_policy" "main" {
  queue_url = aws_sqs_queue.main.id
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.dead_letter.arn
    maxReceiveCount     = 3
  })
}

# Allow DLQ to be used as dead letter queue
resource "aws_sqs_queue_redrive_allow_policy" "dead_letter" {
  queue_url = aws_sqs_queue.dead_letter.id
  redrive_allow_policy = jsonencode({
    redrivePermission = "byQueue",
    sourceQueueArns   = [aws_sqs_queue.main.arn]
  })
}
