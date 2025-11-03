# IAM Role for Lambda Execution
resource "aws_iam_role" "lambda_execution" {
  name               = "${var.project_name}-lambda-execution-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json

  tags = merge(
    var.tags,
    {
      Name = "${var.project_name}-lambda-role"
    }
  )
}

# Assume Role Policy Document
data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

# Lambda Execution Policy
data "aws_iam_policy_document" "lambda_execution" {
  # CloudWatch Logs
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = ["arn:aws:logs:*:*:*"]
  }

  # VPC Network Interfaces (if Lambda is in VPC)
  statement {
    effect = "Allow"
    actions = [
      "ec2:CreateNetworkInterface",
      "ec2:DescribeNetworkInterfaces",
      "ec2:DeleteNetworkInterface",
      "ec2:AssignPrivateIpAddresses",
      "ec2:UnassignPrivateIpAddresses"
    ]
    resources = ["*"]
  }

  # S3 Access
  dynamic "statement" {
    for_each = var.s3_bucket_arns
    content {
      effect = "Allow"
      actions = [
        "s3:GetObject",
        "s3:PutObject",
        "s3:DeleteObject"
      ]
      resources = ["${statement.value}/*"]
    }
  }

  dynamic "statement" {
    for_each = var.s3_bucket_arns
    content {
      effect = "Allow"
      actions = [
        "s3:ListBucket"
      ]
      resources = [statement.value]
    }
  }

  # SQS Access
  dynamic "statement" {
    for_each = var.sqs_queue_arns
    content {
      effect = "Allow"
      actions = [
        "sqs:SendMessage",
        "sqs:ReceiveMessage",
        "sqs:DeleteMessage",
        "sqs:GetQueueAttributes",
        "sqs:ChangeMessageVisibility"
      ]
      resources = [statement.value]
    }
  }

  # SES Access
  dynamic "statement" {
    for_each = var.enable_ses_access ? [1] : []
    content {
      effect = "Allow"
      actions = [
        "ses:SendEmail",
        "ses:SendRawEmail"
      ]
      resources = ["*"]
    }
  }

  # Secrets Manager (optional)
  dynamic "statement" {
    for_each = var.secrets_manager_arns
    content {
      effect = "Allow"
      actions = [
        "secretsmanager:GetSecretValue"
      ]
      resources = [statement.value]
    }
  }

  # KMS (optional)
  dynamic "statement" {
    for_each = var.kms_key_arns
    content {
      effect = "Allow"
      actions = [
        "kms:Decrypt",
        "kms:Encrypt",
        "kms:GenerateDataKey"
      ]
      resources = [statement.value]
    }
  }
}

# Attach Policy to Role
resource "aws_iam_role_policy" "lambda_execution" {
  name   = "${var.project_name}-lambda-execution-policy"
  role   = aws_iam_role.lambda_execution.id
  policy = data.aws_iam_policy_document.lambda_execution.json
}

# Attach AWS Managed Policies (optional)
resource "aws_iam_role_policy_attachment" "lambda_vpc_execution" {
  count = var.attach_vpc_policy ? 1 : 0

  role       = aws_iam_role.lambda_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
}

resource "aws_iam_role_policy_attachment" "lambda_basic_execution" {
  count = var.attach_basic_policy ? 1 : 0

  role       = aws_iam_role.lambda_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}
