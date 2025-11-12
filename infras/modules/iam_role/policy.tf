data "aws_iam_policy_document" "lambda_app_policy" {
  statement {
    sid = "S3UploadAccess"
    actions = [
      "s3:GetObject",
      "s3:PutObject",
      "s3:ListBucket"
    ]
    resources = [
      var.s3_bucket_arn,
      "${var.s3_bucket_arn}/*"
    ]
  }

  statement {
    sid = "SESAccess"
    actions = [
      "ses:SendEmail",
      "ses:SendRawEmail"
    ]
    resources = [var.ses_identity_arn]
  }

  statement {
    sid = "SQSAccess"
    actions = [ 
      "sqs:SendMessage",
      "sqs:ReceiveMessage",
      "sqs:DeleteMessage",
      "sqs:GetQueueAttributes"
    ]
    resources = [
      var.sqs_queue_arn,
      var.sqs_dlq_arn
    ]
  }

  statement {
    sid = "CloudWatchLogsAccess"
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = ["arn:aws:logs:*:*:*"]
  }
}
