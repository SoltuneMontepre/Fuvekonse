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

# SES Variables

variable "ses_sender_email" {
  description = "Email address to verify and use as sender for SES"
  type        = string
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

# Database Variables

variable "db_host" {
  description = "Database host endpoint"
  type        = string
}

variable "db_port" {
  description = "Database port"
  type        = string
  default     = "5432"
}

variable "db_user" {
  description = "Database username"
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
  description = "Database SSL mode (disable, require, verify-ca, verify-full)"
  type        = string
  default     = "disable"
}

# Redis Variables

variable "redis_host" {
  description = "Redis host endpoint"
  type        = string
}

variable "redis_port" {
  description = "Redis port"
  type        = string
  default     = "6379"
}

variable "redis_url" {
  description = "Full Redis connection URL"
  type        = string
}

# Application Variables

variable "jwt_secret" {
  description = "JWT secret key for token generation"
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
  description = "Maximum failed login attempts before blocking"
  type        = string
  default     = "5"
}

variable "login_fail_block_minutes" {
  description = "Duration in minutes to block after max failed attempts"
  type        = string
  default     = "15"
}

variable "frontend_url" {
  description = "Frontend application URL"
  type        = string
}

variable "gin_mode" {
  description = "Gin framework mode (debug or release)"
  type        = string
  default     = "release"
}