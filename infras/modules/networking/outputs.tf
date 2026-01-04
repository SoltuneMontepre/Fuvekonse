output "api_gateway_url" {
  description = "Base URL of the HTTP API Gateway"
  value       = aws_apigatewayv2_api.main.api_endpoint
}

output "api_gateway_id" {
  description = "ID of the HTTP API Gateway"
  value       = aws_apigatewayv2_api.main.id
}

output "general_service_url" {
  description = "Full URL for the general service"
  value       = "${aws_apigatewayv2_api.main.api_endpoint}/api/general"
}

output "rbac_service_url" {
  description = "Full URL for the RBAC service (legacy /api/ticket path for backwards compatibility)"
  value       = "${aws_apigatewayv2_api.main.api_endpoint}/api/ticket"
}
