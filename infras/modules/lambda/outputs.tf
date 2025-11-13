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

output "ticket_service_function_name" {
  description = "Name of the ticket service Lambda function"
  value       = aws_lambda_function.ticket_service.function_name
}

output "ticket_service_function_arn" {
  description = "ARN of the ticket service Lambda function"
  value       = aws_lambda_function.ticket_service.arn
}

output "ticket_service_invoke_arn" {
  description = "Invoke ARN of the ticket service Lambda function"
  value       = aws_lambda_function.ticket_service.invoke_arn
}
