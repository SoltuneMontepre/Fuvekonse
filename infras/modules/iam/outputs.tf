output "read_only_user_arn" {
	description = "ARN of the read-only IAM user created for S3 access"
	value       = aws_iam_user.read_only_user.arn
}

output "lambda_app_user_arn" {
  description = "ARN of the IAM user created for Lambda application"
  value       = aws_iam_user.lambda_app.arn
}

output "lambda_app_user_name" {
  description = "Name of the IAM user created for Lambda application"
  value       = aws_iam_user.lambda_app.name
}

output "read_only_user_name" {
	description = "Name of the read-only IAM user"
	value       = aws_iam_user.read_only_user.name
}

output "read_only_password" {
  description = "Password of the read-only IAM user"
  value       = aws_iam_user_login_profile.read_only_user_login.encrypted_password
}

output "lambda_app_password" {
  description = "Password of the Lambda application IAM user"
  value       = aws_iam_user_login_profile.lambda_app_login.encrypted_password
}

output "lambda_app_access_key_id" {
  description = "Access key ID for the Lambda application IAM user"
  value       = aws_iam_access_key.lambda_app_key.id
}

output "lambda_app_secret_access_key" {
  description = "Secret access key for the Lambda application IAM user (sensitive)"
  value       = aws_iam_access_key.lambda_app_key.secret
  sensitive   = true
}