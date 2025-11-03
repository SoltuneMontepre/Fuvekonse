# General Configuration
variable "aws_region" {
  description = "AWS Region"
  type        = string
  default     = "ap-east-1"
}

variable "project_name" {
  description = "Project name for resource naming"
  type        = string
  default     = "fuvekonse"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "production"
}

# VPC Configuration
variable "vpc_cidr" {
  description = "VPC CIDR block"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "List of availability zones"
  type        = list(string)
  default     = ["ap-east-1a", "ap-east-1b"]
}

variable "public_subnet_cidrs" {
  description = "CIDR blocks for public subnets"
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24"]
}

variable "private_subnet_cidrs" {
  description = "CIDR blocks for private subnets"
  type        = list(string)
  default     = ["10.0.10.0/24", "10.0.11.0/24"]
}

# Database Configuration
variable "db_name" {
  description = "Database name"
  type        = string
  default     = "fuvekonse_db"
}

variable "db_username" {
  description = "Database master username"
  type        = string
  sensitive   = true
}

variable "db_password" {
  description = "Database master password"
  type        = string
  sensitive   = true
}

variable "db_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t3.micro"
}

variable "db_allocated_storage" {
  description = "Allocated storage in GB"
  type        = number
  default     = 20
}

variable "db_engine_version" {
  description = "PostgreSQL engine version"
  type        = string
  default     = "15.4"
}

# ElastiCache Configuration
variable "redis_node_type" {
  description = "ElastiCache node type"
  type        = string
  default     = "cache.t3.micro"
}

variable "redis_num_cache_nodes" {
  description = "Number of cache nodes"
  type        = number
  default     = 2
}

variable "redis_engine_version" {
  description = "Redis engine version"
  type        = string
  default     = "7.0"
}

# Lambda Configuration
variable "lambda_timeout" {
  description = "Lambda function timeout in seconds"
  type        = number
  default     = 30
}

variable "lambda_memory_size" {
  description = "Lambda function memory in MB"
  type        = number
  default     = 256
}

# S3 Configuration
variable "bucket_name" {
  description = "S3 bucket name"
  type        = string
  default     = "fuvekonse-s3-bucket"
}

# SQS Configuration
variable "sqs_visibility_timeout" {
  description = "SQS visibility timeout in seconds"
  type        = number
  default     = 30
}

# Domain Configuration
variable "domain_name" {
  description = "Domain name for SES"
  type        = string
  default     = "fuvekonse.com"
}

# Common Tags
variable "tags" {
  description = "Common tags for all resources"
  type        = map(string)
  default     = {}
}
