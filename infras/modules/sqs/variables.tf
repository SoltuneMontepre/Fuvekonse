variable "name" {
  description = "Name of the SQS queue"
  type        = string
}

variable "delay_seconds" {
  description = "Delay in seconds for message delivery"
  type        = number
  default     = 0
}

variable "max_message_size" {
  description = "Maximum message size in bytes"
  type        = number
  default     = 262144
}

variable "message_retention_seconds" {
  description = "Message retention period in seconds"
  type        = number
  default     = 345600 # 4 days
}

variable "receive_wait_time_seconds" {
  description = "Long polling wait time in seconds"
  type        = number
  default     = 0
}

variable "visibility_timeout_seconds" {
  description = "Visibility timeout in seconds"
  type        = number
  default     = 30
}

variable "fifo_queue" {
  description = "Create FIFO queue"
  type        = bool
  default     = false
}

variable "content_based_deduplication" {
  description = "Enable content-based deduplication (FIFO only)"
  type        = bool
  default     = false
}

variable "kms_master_key_id" {
  description = "KMS key ID for encryption"
  type        = string
  default     = null
}

variable "kms_data_key_reuse_period_seconds" {
  description = "KMS data key reuse period in seconds"
  type        = number
  default     = 300
}

variable "create_dlq" {
  description = "Create dead letter queue"
  type        = bool
  default     = false
}

variable "dlq_message_retention_seconds" {
  description = "DLQ message retention period in seconds"
  type        = number
  default     = 1209600 # 14 days
}

variable "max_receive_count" {
  description = "Max receive count before sending to DLQ"
  type        = number
  default     = 5
}

variable "queue_policy" {
  description = "JSON policy for the queue"
  type        = string
  default     = null
}

variable "tags" {
  description = "Additional tags for the queue"
  type        = map(string)
  default     = {}
}
