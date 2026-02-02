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

variable "doppler_project" {
  description = "Doppler project name"
  type        = string
}

variable "doppler_config" {
  description = "Doppler config name (e.g., dev, stg, prd)"
  type        = string
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

variable "iam_s3_upload_username" {
  description = "IAM username for S3 upload access from frontend"
  type        = string
  default     = "fuvekon-s3-upload"
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

variable "rbac_service_zip_path" {
  description = "Path to the RBAC service deployment package (zip file)"
  type        = string
}

variable "sqs_worker_zip_path" {
  description = "Path to the sqs-worker service deployment package (zip file)"
  type        = string
}

variable "gin_mode" {
  description = "Gin framework mode (debug or release)"
  type        = string
  default     = "release"
}

variable "general_service_url" {
  description = "Base URL of general-service API (for sqs-worker to call /internal/jobs/ticket)"
  type        = string
  default     = ""
}

variable "internal_api_key" {
  description = "Internal API key for general-service /internal/jobs/ticket"
  type        = string
  default     = ""
  sensitive   = true
}
