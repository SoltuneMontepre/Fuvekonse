resource "aws_iam_role" "lambda_app_role" {
  name = var.iam_lambda_app_role_name

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name        = var.project_name
    Environment = "Production"
  }
}

resource "aws_iam_role_policy" "lambda_app_inline" {
  name   = "lambda-app-inline-policy"
  role   = aws_iam_role.lambda_app_role.id
  policy = data.aws_iam_policy_document.lambda_app_policy.json
}
