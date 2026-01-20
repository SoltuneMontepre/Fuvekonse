output "general_service_function_name" {
  description = "Name of the general service Lambda function"
  value       = aws_lambda_function.general_service.function_name
}

output "general_service_function_arn" {
  description = "ARN of the general service Lambda function"
  value       = aws_lambda_function.general_service.arn
}

output "general_service_invoke_arn" {
  description = "Invoke ARN of the general service Lambda function"
  value       = aws_lambda_function.general_service.invoke_arn
}

output "rbac_service_function_name" {
  description = "Name of the RBAC service Lambda function"
  value       = aws_lambda_function.rbac_service.function_name
}

output "rbac_service_function_arn" {
  description = "ARN of the RBAC service Lambda function"
  value       = aws_lambda_function.rbac_service.arn
}

output "rbac_service_invoke_arn" {
  description = "Invoke ARN of the RBAC service Lambda function"
  value       = aws_lambda_function.rbac_service.invoke_arn
}

output "sqs_worker_function_name" {
  description = "Name of the SQS worker Lambda function"
  value       = aws_lambda_function.sqs_worker.function_name
}

output "sqs_worker_function_arn" {
  description = "ARN of the SQS worker Lambda function"
  value       = aws_lambda_function.sqs_worker.arn
}
