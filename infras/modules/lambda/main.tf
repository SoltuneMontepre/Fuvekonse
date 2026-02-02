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
      DB_HOST                         = var.db_host
      DB_PORT                         = var.db_port
      DB_USER                         = var.db_user
      DB_PASSWORD                     = var.db_password
      DB_NAME                         = var.db_name
      DB_SSLMODE                      = var.db_sslmode
      REDIS_URL                       = var.redis_url
      REDIS_TLS                       = "true"
      JWT_SECRET                      = var.jwt_secret
      JWT_ACCESS_TOKEN_EXPIRY_MINUTES = var.jwt_access_token_expiry_minutes
      JWT_REFRESH_TOKEN_EXPIRY_DAYS   = var.jwt_refresh_token_expiry_days
      LOGIN_MAX_FAIL                  = var.login_max_fail
      LOGIN_FAIL_BLOCK_MINUTES        = var.login_fail_block_minutes
      FRONTEND_URL                    = var.frontend_url
      CORS_ALLOWED_ORIGINS            = join(",", var.cors_allowed_origins)
      GIN_MODE                        = var.gin_mode
      S3_BUCKET                       = var.s3_bucket_name
      SES_SENDER                      = var.ses_sender_email
      SQS_QUEUE                       = var.sqs_queue_url
      INTERNAL_API_KEY                = var.internal_api_key
      COOKIE_DOMAIN                   = ""
      COOKIE_SECURE                   = "true"
      COOKIE_SAMESITE                 = "None"
      INTERNAL_API_KEY                = var.internal_api_key
    }
  }

  tags = {
    Name        = var.project_name
    Environment = "Production"
    Service     = "general-service"
  }
}

# RBAC Service Lambda Function
resource "aws_lambda_function" "rbac_service" {
  function_name = "${var.project_name}-rbac-service"
  role          = var.lambda_role_arn
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  timeout       = 30
  memory_size   = 512

  filename         = var.rbac_service_zip_path
  source_code_hash = filebase64sha256(var.rbac_service_zip_path)

  environment {
    variables = {
      DB_HOST                         = var.db_host
      DB_PORT                         = var.db_port
      DB_USER                         = var.db_user
      DB_PASSWORD                     = var.db_password
      DB_NAME                         = var.db_name
      DB_SSLMODE                      = var.db_sslmode
      REDIS_URL                       = var.redis_url
      REDIS_TLS                       = "true"
      JWT_SECRET                      = var.jwt_secret
      JWT_ACCESS_TOKEN_EXPIRY_MINUTES = var.jwt_access_token_expiry_minutes
      JWT_REFRESH_TOKEN_EXPIRY_DAYS   = var.jwt_refresh_token_expiry_days
      LOGIN_MAX_FAIL                  = var.login_max_fail
      LOGIN_FAIL_BLOCK_MINUTES        = var.login_fail_block_minutes
      FRONTEND_URL                    = var.frontend_url
      CORS_ALLOWED_ORIGINS            = join(",", var.cors_allowed_origins)
      GIN_MODE                        = var.gin_mode
      S3_BUCKET                       = var.s3_bucket_name
      SES_SENDER                      = var.ses_sender_email
      SQS_QUEUE                       = var.sqs_queue_url
      COOKIE_DOMAIN                   = ""
      COOKIE_SECURE                   = "true"
      COOKIE_SAMESITE                 = "None"
      INTERNAL_API_KEY                = var.internal_api_key
    }
  }

  tags = {
    Name        = var.project_name
    Environment = "Production"
    Service     = "rbac-service"
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

resource "aws_cloudwatch_log_group" "rbac_service" {
  name              = "/aws/lambda/${aws_lambda_function.rbac_service.function_name}"
  retention_in_days = 14

  tags = {
    Name        = var.project_name
    Environment = "Production"
  }
}

# SQS Worker Lambda Function
resource "aws_lambda_function" "sqs_worker" {
  function_name = "${var.project_name}-sqs-worker"
  role          = var.lambda_role_arn
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  timeout       = 60
  memory_size   = 256

  filename         = var.sqs_worker_zip_path
  source_code_hash = filebase64sha256(var.sqs_worker_zip_path)

  environment {
    variables = {
      DB_HOST              = var.db_host
      DB_PORT              = var.db_port
      DB_USER              = var.db_user
      DB_PASSWORD          = var.db_password
      DB_NAME              = var.db_name
      DB_SSLMODE           = var.db_sslmode
      SES_SENDER           = var.ses_sender_email
      SQS_QUEUE            = var.sqs_queue_url
      GENERAL_SERVICE_URL  = var.general_service_url
      INTERNAL_API_KEY     = var.internal_api_key
    }
  }

  tags = {
    Name        = var.project_name
    Environment = "Production"
    Service     = "sqs-worker"
  }
}

# CloudWatch Log Group for SQS Worker
resource "aws_cloudwatch_log_group" "sqs_worker" {
  name              = "/aws/lambda/${aws_lambda_function.sqs_worker.function_name}"
  retention_in_days = 14

  tags = {
    Name        = var.project_name
    Environment = "Production"
  }
}

# Event Source Mapping - Trigger Lambda from SQS
resource "aws_lambda_event_source_mapping" "sqs_worker" {
  event_source_arn = var.sqs_queue_arn
  function_name    = aws_lambda_function.sqs_worker.arn
  batch_size       = 10
  enabled          = true

  function_response_types = ["ReportBatchItemFailures"]

  scaling_config {
    maximum_concurrency = 2
  }
}
