variable "bucket_name" {
  description = "Name of the S3 for fuvekon"
  type        = string
}
variable "bucket_acl" {
  description = "Access control list for S3 bucket"
  type        = string
  default     = "private"
}

variable "project_name" {
  type = string
  default = "fuvekon"
}

variable "read_only_principal_arns" {
  description = "List of IAM principal ARNs allowed to read objects (user or role ARNs). If empty, no principals will be added to the bucket policy."
  type        = list(string)
  default     = []
}