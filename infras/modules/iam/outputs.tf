output "read_only_user_arn" {
	description = "ARN of the read-only IAM user created for S3 access"
	value       = aws_iam_user.read_only_user.arn
}

output "read_only_user_name" {
	description = "Name of the read-only IAM user"
	value       = aws_iam_user.read_only_user.name
}

output "read_only_password" {
  description = "Password of the read-only IAM user"
  value       = aws_iam_user_login_profile.read_only_user_login.encrypted_password
}

output "s3_upload_user_name" {
  description = "Name of the S3 upload IAM user"
  value       = aws_iam_user.s3_upload_user.name
}

output "s3_upload_access_key_id" {
  description = "Access key ID for S3 upload user"
  value       = aws_iam_access_key.s3_upload_user_key.id
}

output "s3_upload_secret_access_key" {
  description = "Secret access key for S3 upload user"
  value       = aws_iam_access_key.s3_upload_user_key.secret
  sensitive   = true
}
