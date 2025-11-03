variable "project_name" {
  description = "Project name for resource naming"
  type        = string
}

variable "s3_bucket_arns" {
  description = "List of S3 bucket ARNs to grant access to"
  type        = list(string)
  default     = []
}

variable "sqs_queue_arns" {
  description = "List of SQS queue ARNs to grant access to"
  type        = list(string)
  default     = []
}

variable "enable_ses_access" {
  description = "Enable SES access"
  type        = bool
  default     = false
}

variable "secrets_manager_arns" {
  description = "List of Secrets Manager ARNs to grant access to"
  type        = list(string)
  default     = []
}

variable "kms_key_arns" {
  description = "List of KMS key ARNs to grant access to"
  type        = list(string)
  default     = []
}

variable "attach_vpc_policy" {
  description = "Attach AWS managed VPC execution policy"
  type        = bool
  default     = true
}

variable "attach_basic_policy" {
  description = "Attach AWS managed basic execution policy"
  type        = bool
  default     = false
}

variable "tags" {
  description = "Additional tags for IAM resources"
  type        = map(string)
  default     = {}
}
