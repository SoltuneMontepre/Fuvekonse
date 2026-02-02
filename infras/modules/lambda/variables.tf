variable "project_name" {
  description = "The name of the project"
  type        = string
}

variable "lambda_role_arn" {
  description = "ARN of the IAM role for Lambda functions"
  type        = string
}

variable "general_service_zip_path" {
  description = "Path to the general service deployment package"
  type        = string
}

variable "rbac_service_zip_path" {
  description = "Path to the RBAC service deployment package"
  type        = string
}

variable "sqs_worker_zip_path" {
  description = "Path to the sqs-worker service deployment package"
  type        = string
}

variable "db_host" {
  description = "Database host"
  type        = string
}

variable "db_port" {
  description = "Database port"
  type        = string
  default     = "5432"
}

variable "db_user" {
  description = "Database user"
  type        = string
}

variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true
}

variable "db_name" {
  description = "Database name"
  type        = string
}

variable "db_sslmode" {
  description = "Database SSL mode"
  type        = string
  default     = "disable"
}

variable "redis_url" {
  description = "Full Redis URL"
  type        = string
}

variable "jwt_secret" {
  description = "JWT secret key"
  type        = string
  sensitive   = true
}

variable "jwt_access_token_expiry_minutes" {
  description = "JWT access token expiry in minutes"
  type        = string
  default     = "15"
}

variable "jwt_refresh_token_expiry_days" {
  description = "JWT refresh token expiry in days"
  type        = string
  default     = "7"
}

variable "login_max_fail" {
  description = "Maximum failed login attempts"
  type        = string
  default     = "5"
}

variable "login_fail_block_minutes" {
  description = "Login fail block duration in minutes"
  type        = string
  default     = "15"
}

variable "frontend_url" {
  description = "Frontend URL for CORS"
  type        = string
}

variable "gin_mode" {
  description = "Gin framework mode (debug or release)"
  type        = string
  default     = "release"
}

variable "aws_region" {
  description = "AWS region"
  type        = string
}

variable "s3_bucket_name" {
  description = "S3 bucket name"
  type        = string
}

variable "ses_sender_email" {
  description = "SES sender email address"
  type        = string
}

variable "sqs_queue_url" {
  description = "SQS queue URL"
  type        = string
}

variable "cors_allowed_origins" {
  description = "List of allowed origins for CORS"
  type        = list(string)
  default     = []
}

variable "sqs_queue_arn" {
  description = "SQS queue ARN for event source mapping"
  type        = string
}

variable "internal_api_key" {
  description = "Internal API key for service-to-service communication"
  type        = string
  sensitive   = true
}

variable "general_service_url" {
  description = "URL of the general service for internal API calls"
  type        = string
}
