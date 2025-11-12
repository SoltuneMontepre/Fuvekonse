# Provider Variables

variable "aws_region" {
  description = "AWS region for deploying fuvekon resources"
  type        = string
  default     = "ap-southeast-1"
}

variable "project_name" {
  type    = string
  default = "fuvekon"
}

# Doppler Variable
variable "doppler_token" {
  description = "Doppler service token for accessing secrets"
  type        = string
  sensitive   = true
}


# S3 Bucket Variables
variable "bucket_name" {
  description = "Name of the S3 for fuvekon"
  type        = string
}

variable "bucket_acl" {
  description = "Access control list for S3 bucket"
  type        = string
  default     = "private"
}

# IAM Variables

variable "iam_bucket_access_username" {
  description = "List of IAM user ARNs to grant read/list access to the S3 bucket"
  type        = string
}

variable "iam_lambda_app_role_name" {
  description = "The name of the IAM role for Lambda application"
  type        = string
}

variable "s3_cors_allowed_origins" {
  description = "List of allowed origins for CORS configuration"
  type        = list(string)
  default     = []
}

# Lambda Variables

variable "general_service_zip_path" {
  description = "Path to the general service deployment package (zip file)"
  type        = string
}

variable "ticket_service_zip_path" {
  description = "Path to the ticket service deployment package (zip file)"
  type        = string
}

variable "gin_mode" {
  description = "Gin framework mode (debug or release)"
  type        = string
  default     = "release"
}

# NOTE: The following variables are now managed by Doppler and fetched via doppler.tf
# - db_host, db_port, db_user, db_password, db_name, db_sslmode
# - redis_host, redis_port, redis_url
# - jwt_secret, jwt_access_token_expiry_minutes, jwt_refresh_token_expiry_days
# - login_max_fail, login_fail_block_minutes
# - frontend_url
# - ses_sender_email
