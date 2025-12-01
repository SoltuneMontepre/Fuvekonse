variable "iam_bucket_access_username" {
  description = "The name of the IAM user to access the S3 bucket"
  type        = string
}

variable "iam_s3_upload_username" {
  description = "The name of the IAM user for S3 uploads from frontend"
  type        = string
}

variable "project_name" {
  description = "The name of the project"
  type        = string
}

variable "s3_bucket_name" {
  description = "The name of the S3 bucket"
  type        = string
}
