variable "function_name" {
  description = "Name of the Lambda function"
  type        = string
}

variable "filename" {
  description = "Path to the deployment package"
  type        = string
  default     = null
}

variable "role_arn" {
  description = "ARN of the IAM role for Lambda execution"
  type        = string
}

variable "handler" {
  description = "Function entrypoint"
  type        = string
}

variable "source_code_hash" {
  description = "Base64-encoded SHA256 hash of the package"
  type        = string
  default     = null
}

variable "runtime" {
  description = "Runtime identifier"
  type        = string
}

variable "timeout" {
  description = "Function timeout in seconds"
  type        = number
  default     = 30
}

variable "memory_size" {
  description = "Amount of memory in MB"
  type        = number
  default     = 128
}

variable "publish" {
  description = "Publish a version"
  type        = bool
  default     = false
}

variable "layers" {
  description = "List of Lambda Layer ARNs"
  type        = list(string)
  default     = []
}

variable "architectures" {
  description = "Instruction set architectures (x86_64 or arm64)"
  type        = list(string)
  default     = ["x86_64"]
}

variable "vpc_config" {
  description = "VPC configuration for the Lambda function"
  type = object({
    subnet_ids         = list(string)
    security_group_ids = list(string)
  })
  default = null
}

variable "environment_variables" {
  description = "Environment variables"
  type        = map(string)
  default     = null
  sensitive   = true
}

variable "dead_letter_config" {
  description = "Dead letter queue configuration"
  type = object({
    target_arn = string
  })
  default = null
}

variable "tracing_mode" {
  description = "X-Ray tracing mode (Active or PassThrough)"
  type        = string
  default     = null
}

variable "reserved_concurrent_executions" {
  description = "Reserved concurrent executions"
  type        = number
  default     = -1
}

variable "create_log_group" {
  description = "Create CloudWatch Log Group"
  type        = bool
  default     = true
}

variable "log_retention_days" {
  description = "Log retention in days"
  type        = number
  default     = 7
}

variable "log_kms_key_id" {
  description = "KMS key ID for log encryption"
  type        = string
  default     = null
}

variable "create_function_url" {
  description = "Create Lambda Function URL"
  type        = bool
  default     = false
}

variable "function_url_auth_type" {
  description = "Authorization type for Function URL (AWS_IAM or NONE)"
  type        = string
  default     = "AWS_IAM"
}

variable "function_url_cors" {
  description = "CORS configuration for Function URL"
  type        = any
  default     = null
}

variable "event_source_arn" {
  description = "ARN of the event source (SQS, Kinesis, DynamoDB)"
  type        = string
  default     = null
}

variable "event_source_batch_size" {
  description = "Batch size for event source"
  type        = number
  default     = 10
}

variable "event_source_starting_position" {
  description = "Starting position for stream sources (LATEST or TRIM_HORIZON)"
  type        = string
  default     = "LATEST"
}

variable "event_source_filter_criteria" {
  description = "Filter criteria for event source"
  type        = any
  default     = null
}

variable "tags" {
  description = "Additional tags for the function"
  type        = map(string)
  default     = {}
}
