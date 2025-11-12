variable "iam_lambda_app_role_name" {
  description = "The name of the IAM role for Lambda application"
  type        = string
}

variable "project_name" {
  description = "The name of the project"
  type        = string
}

variable "s3_bucket_arn" {
  description = "ARN of the S3 bucket"
  type        = string
}

variable "ses_identity_arn" {
  description = "ARN of the SES email identity"
  type        = string
}

variable "sqs_queue_arn" {
  description = "ARN of the SQS queue"
  type        = string
}

variable "sqs_dlq_arn" {
  description = "ARN of the SQS dead letter queue"
  type        = string
}
