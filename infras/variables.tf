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

variable "iam_bucket_access_username" {
  description = "List of IAM user ARNs to grant read/list access to the S3 bucket"
  type        = string
}