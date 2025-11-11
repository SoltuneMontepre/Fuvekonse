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