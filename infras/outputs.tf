# S3 Outputs
output "s3_bucket_name" {
  description = "Name of the S3 bucket"
  value       = module.s3.bucket_name
}

output "s3_bucket_arn" {
  description = "ARN of the S3 bucket"
  value       = module.s3.bucket_arn
}

# IAM Outputs
output "read_only_user_name" {
  description = "Name of the read-only IAM user"
  value       = module.iam.read_only_user_name
}

output "read_only_user_password" {
  description = "Password of the read-only IAM user"
  value       = module.iam.read_only_password
  sensitive   = true
}

output "lambda_app_role_arn" {
  description = "ARN of the Lambda application IAM role"
  value       = module.iam_role.lambda_app_role_arn
}

# SES Outputs
output "ses_sender_email" {
  description = "Verified SES sender email"
  value       = module.ses.sender_email
  sensitive   = true
}

# SQS Outputs
output "sqs_queue_url" {
  description = "URL of the SQS queue"
  value       = module.sqs.queue_url
}

output "sqs_queue_arn" {
  description = "ARN of the SQS queue"
  value       = module.sqs.queue_arn
}

output "general_service_url" {
  description = "HTTPS URL for the general service (use this in your frontend)"
  value       = module.networking.general_service_url
}

output "rbac_service_url" {
  description = "HTTPS URL for the RBAC service (legacy /api/ticket path for backwards compatibility)"
  value       = module.networking.rbac_service_url
}

output "api_gateway_url" {
  description = "Base URL of the HTTP API Gateway"
  value       = module.networking.api_gateway_url
}

output "api_gateway_id" {
  description = "ID of the HTTP API Gateway"
  value       = module.networking.api_gateway_id
}

output "general_service_function_name" {
  description = "Name of the general service Lambda function"
  value       = module.lambda.general_service_function_name
}

output "rbac_service_function_name" {
  description = "Name of the RBAC service Lambda function"
  value       = module.lambda.rbac_service_function_name
}

output "sqs_worker_function_name" {
  description = "Name of the SQS worker Lambda function"
  value       = module.lambda.sqs_worker_function_name
}

# S3 Upload User Outputs (for Frontend)
output "s3_upload_access_key_id" {
  description = "Access Key ID for S3 uploads from frontend"
  value       = module.iam.s3_upload_access_key_id
}

output "s3_upload_secret_access_key" {
  description = "Secret Access Key for S3 uploads from frontend"
  value       = module.iam.s3_upload_secret_access_key
  sensitive   = true
}

output "s3_upload_user_name" {
  description = "Username of the S3 upload IAM user"
  value       = module.iam.s3_upload_user_name
}

output "sqs_worker_function_arn" {
  description = "ARN of the SQS worker Lambda function"
  value       = module.lambda.sqs_worker_function_arn
}
