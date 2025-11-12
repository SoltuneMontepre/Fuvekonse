output "lambda_app_role_arn" {
  description = "ARN of the IAM role created for Lambda application"
  value       = aws_iam_role.lambda_app_role.arn
}

output "lambda_app_role_name" {
  description = "Name of the IAM role created for Lambda application"
  value       = aws_iam_role.lambda_app_role.name
}
