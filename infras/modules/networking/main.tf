# HTTP API Gateway
resource "aws_apigatewayv2_api" "main" {
  name                         = "${var.project_name}-api"
  protocol_type                = "HTTP"
  disable_execute_api_endpoint = true

  cors_configuration {
    allow_origins     = var.cors_allowed_origins
    allow_methods     = ["GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"]
    allow_headers     = ["content-type", "authorization", "x-amz-date", "x-api-key", "x-amz-security-token"]
    expose_headers    = ["set-cookie"]
    allow_credentials = true
    max_age           = 300
  }

  tags = {
    Name        = var.project_name
    Environment = "Production"
  }
}

# Default stage with auto-deploy
resource "aws_apigatewayv2_stage" "default" {
  api_id      = aws_apigatewayv2_api.main.id
  name        = "$default"
  auto_deploy = true

  tags = {
    Name        = var.project_name
    Environment = "Production"
  }
}

# General Service Integration
resource "aws_apigatewayv2_integration" "general_service" {
  api_id                 = aws_apigatewayv2_api.main.id
  integration_type       = "AWS_PROXY"
  integration_uri        = var.general_service_invoke_arn
  payload_format_version = "2.0"
  timeout_milliseconds   = 30000
}

# General Service Routes - catch all paths
resource "aws_apigatewayv2_route" "general_service" {
  api_id             = aws_apigatewayv2_api.main.id
  route_key          = "ANY /api/general/{proxy+}"
  target             = "integrations/${aws_apigatewayv2_integration.general_service.id}"
  authorization_type = "NONE"
}

# General Service root route
resource "aws_apigatewayv2_route" "general_service_root" {
  api_id             = aws_apigatewayv2_api.main.id
  route_key          = "ANY /api/general"
  target             = "integrations/${aws_apigatewayv2_integration.general_service.id}"
  authorization_type = "NONE"
}

# Lambda permission for API Gateway to invoke general service
resource "aws_lambda_permission" "general_service_api" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = var.general_service_function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.main.execution_arn}/*/*"
}

# RBAC Service Integration (serves /api/ticket endpoints)
resource "aws_apigatewayv2_integration" "rbac_service" {
  api_id                 = aws_apigatewayv2_api.main.id
  integration_type       = "AWS_PROXY"
  integration_uri        = var.rbac_service_invoke_arn
  payload_format_version = "2.0"
  timeout_milliseconds   = 30000
}

# RBAC Service Routes - catch all paths (legacy /api/ticket path for backwards compatibility)
resource "aws_apigatewayv2_route" "rbac_service" {
  api_id             = aws_apigatewayv2_api.main.id
  route_key          = "ANY /api/ticket/{proxy+}"
  target             = "integrations/${aws_apigatewayv2_integration.rbac_service.id}"
  authorization_type = "NONE"
}

# RBAC Service root route (legacy /api/ticket path for backwards compatibility)
resource "aws_apigatewayv2_route" "rbac_service_root" {
  api_id             = aws_apigatewayv2_api.main.id
  route_key          = "ANY /api/ticket"
  target             = "integrations/${aws_apigatewayv2_integration.rbac_service.id}"
  authorization_type = "NONE"
}

# Lambda permission for API Gateway to invoke RBAC service
resource "aws_lambda_permission" "rbac_service_api" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = var.rbac_service_function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.main.execution_arn}/*/*"
}
