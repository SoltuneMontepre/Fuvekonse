# General Service Lambda Function
resource "aws_lambda_function" "general_service" {
  function_name = "${var.project_name}-general-service"
  role          = var.lambda_role_arn
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  timeout       = 30
  memory_size   = 512

  filename         = var.general_service_zip_path
  source_code_hash = filebase64sha256(var.general_service_zip_path)

  environment {
    variables = {
      DB_HOST                          = var.db_host
      DB_PORT                          = var.db_port
      DB_USER                          = var.db_user
      DB_PASSWORD                      = var.db_password
      DB_NAME                          = var.db_name
      DB_SSLMODE                       = var.db_sslmode
      REDIS_URL                        = var.redis_url
      REDIS_TLS                        = "true"
      JWT_SECRET                       = var.jwt_secret
      JWT_ACCESS_TOKEN_EXPIRY_MINUTES  = var.jwt_access_token_expiry_minutes
      JWT_REFRESH_TOKEN_EXPIRY_DAYS    = var.jwt_refresh_token_expiry_days
      LOGIN_MAX_FAIL                   = var.login_max_fail
      LOGIN_FAIL_BLOCK_MINUTES         = var.login_fail_block_minutes
      FRONTEND_URL                     = var.frontend_url
      CORS_ALLOWED_ORIGINS             = join(",", var.cors_allowed_origins)
      GIN_MODE                         = var.gin_mode
      S3_BUCKET                        = var.s3_bucket_name
      SES_SENDER                       = var.ses_sender_email
      SQS_QUEUE                        = var.sqs_queue_url
    }
  }

  tags = {
    Name        = var.project_name
    Environment = "Production"
    Service     = "general-service"
  }
}

# Ticket Service Lambda Function
resource "aws_lambda_function" "ticket_service" {
  function_name = "${var.project_name}-ticket-service"
  role          = var.lambda_role_arn
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  timeout       = 30
  memory_size   = 512

  filename         = var.ticket_service_zip_path
  source_code_hash = filebase64sha256(var.ticket_service_zip_path)

  environment {
    variables = {
      DB_HOST                          = var.db_host
      DB_PORT                          = var.db_port
      DB_USER                          = var.db_user
      DB_PASSWORD                      = var.db_password
      DB_NAME                          = var.db_name
      DB_SSLMODE                       = var.db_sslmode
      REDIS_URL                        = var.redis_url
      REDIS_TLS                        = "true"
      JWT_SECRET                       = var.jwt_secret
      JWT_ACCESS_TOKEN_EXPIRY_MINUTES  = var.jwt_access_token_expiry_minutes
      JWT_REFRESH_TOKEN_EXPIRY_DAYS    = var.jwt_refresh_token_expiry_days
      LOGIN_MAX_FAIL                   = var.login_max_fail
      LOGIN_FAIL_BLOCK_MINUTES         = var.login_fail_block_minutes
      FRONTEND_URL                     = var.frontend_url
      CORS_ALLOWED_ORIGINS             = join(",", var.cors_allowed_origins)
      GIN_MODE                         = var.gin_mode
      S3_BUCKET                        = var.s3_bucket_name
      SES_SENDER                       = var.ses_sender_email
      SQS_QUEUE                        = var.sqs_queue_url
    }
  }

  tags = {
    Name        = var.project_name
    Environment = "Production"
    Service     = "ticket-service"
  }
}

# CloudWatch Log Groups
resource "aws_cloudwatch_log_group" "general_service" {
  name              = "/aws/lambda/${aws_lambda_function.general_service.function_name}"
  retention_in_days = 14

  tags = {
    Name        = var.project_name
    Environment = "Production"
  }
}

resource "aws_cloudwatch_log_group" "ticket_service" {
  name              = "/aws/lambda/${aws_lambda_function.ticket_service.function_name}"
  retention_in_days = 14

  tags = {
    Name        = var.project_name
    Environment = "Production"
  }
}

# Lambda Function URL for General Service
resource "aws_lambda_function_url" "general_service" {
  function_name      = aws_lambda_function.general_service.function_name
  authorization_type = "NONE"
  invoke_mode        = "BUFFERED"
}

# Permission for Function URL to invoke General Service Lambda
resource "aws_lambda_permission" "general_service_url" {
  statement_id           = "AllowFunctionURLInvoke"
  action                 = "lambda:InvokeFunctionUrl"
  function_name          = aws_lambda_function.general_service.function_name
  principal              = "*"
  function_url_auth_type = "NONE"
}

# Lambda Function URL for Ticket Service
resource "aws_lambda_function_url" "ticket_service" {
  function_name      = aws_lambda_function.ticket_service.function_name
  authorization_type = "NONE"

  # CORS is handled by the application middleware, not by Lambda Function URL
  # This prevents double CORS header issues
}

# Permission for Function URL to invoke Ticket Service Lambda
resource "aws_lambda_permission" "ticket_service_url" {
  statement_id           = "AllowFunctionURLInvoke"
  action                 = "lambda:InvokeFunctionUrl"
  function_name          = aws_lambda_function.ticket_service.function_name
  principal              = "*"
  function_url_auth_type = "NONE"
}
