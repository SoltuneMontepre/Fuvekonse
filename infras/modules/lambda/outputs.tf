output "general_service_function_name" {
  description = "Name of the general service Lambda function"
  value       = aws_lambda_function.general_service.function_name
}

output "general_service_function_arn" {
  description = "ARN of the general service Lambda function"
  value       = aws_lambda_function.general_service.arn
}

output "general_service_url" {
  description = "HTTPS URL for the general service"
  value       = aws_lambda_function_url.general_service.function_url
}

output "ticket_service_function_name" {
  description = "Name of the ticket service Lambda function"
  value       = aws_lambda_function.ticket_service.function_name
}

output "ticket_service_function_arn" {
  description = "ARN of the ticket service Lambda function"
  value       = aws_lambda_function.ticket_service.arn
}

output "ticket_service_url" {
  description = "HTTPS URL for the ticket service"
  value       = aws_lambda_function_url.ticket_service.function_url
}
