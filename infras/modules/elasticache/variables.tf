variable "name_prefix" {
  description = "Prefix for resource names"
  type        = string
}

variable "replication_group_id" {
  description = "Replication group identifier"
  type        = string
}

variable "description" {
  description = "Description for the replication group"
  type        = string
}

variable "engine_version" {
  description = "Redis engine version"
  type        = string
  default     = "7.0"
}

variable "node_type" {
  description = "Instance type for cache nodes"
  type        = string
  default     = "cache.t3.micro"
}

variable "num_cache_clusters" {
  description = "Number of cache clusters (replicas + 1)"
  type        = number
  default     = 2
}

variable "parameter_group_name" {
  description = "Name of parameter group"
  type        = string
  default     = "default.redis7"
}

variable "port" {
  description = "Port number for Redis"
  type        = number
  default     = 6379
}

variable "subnet_ids" {
  description = "List of subnet IDs for cache subnet group"
  type        = list(string)
}

variable "security_group_ids" {
  description = "List of security group IDs"
  type        = list(string)
}

variable "automatic_failover_enabled" {
  description = "Enable automatic failover"
  type        = bool
  default     = true
}

variable "multi_az_enabled" {
  description = "Enable Multi-AZ"
  type        = bool
  default     = true
}

variable "at_rest_encryption_enabled" {
  description = "Enable encryption at rest"
  type        = bool
  default     = true
}

variable "transit_encryption_enabled" {
  description = "Enable encryption in transit"
  type        = bool
  default     = true
}

variable "auth_token_enabled" {
  description = "Enable auth token (requires transit encryption)"
  type        = bool
  default     = false
}

variable "auth_token" {
  description = "Auth token for Redis (min 16 chars)"
  type        = string
  default     = null
  sensitive   = true
}

variable "snapshot_retention_limit" {
  description = "Number of days to retain backups"
  type        = number
  default     = 5
}

variable "snapshot_window" {
  description = "Daily time range for snapshots"
  type        = string
  default     = "03:00-05:00"
}

variable "maintenance_window" {
  description = "Weekly time range for maintenance"
  type        = string
  default     = "sun:05:00-sun:07:00"
}

variable "notification_topic_arn" {
  description = "ARN of SNS topic for notifications"
  type        = string
  default     = null
}

variable "apply_immediately" {
  description = "Apply changes immediately"
  type        = bool
  default     = false
}

variable "slow_log_destination" {
  description = "Destination for slow log (CloudWatch log group or Kinesis Firehose)"
  type        = string
  default     = null
}

variable "slow_log_destination_type" {
  description = "Destination type for slow log (cloudwatch-logs or kinesis-firehose)"
  type        = string
  default     = "cloudwatch-logs"
}

variable "engine_log_destination" {
  description = "Destination for engine log"
  type        = string
  default     = null
}

variable "engine_log_destination_type" {
  description = "Destination type for engine log"
  type        = string
  default     = "cloudwatch-logs"
}

variable "log_format" {
  description = "Log format (text or json)"
  type        = string
  default     = "text"
}

variable "tags" {
  description = "Additional tags for resources"
  type        = map(string)
  default     = {}
}
