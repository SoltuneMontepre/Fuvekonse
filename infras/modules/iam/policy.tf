data "aws_iam_policy_document" "read_only_user_policy" {
  statement {
    sid = "S3Read"
    actions = [
      "s3:GetObject",
      "s3:ListBucket",
      "s3:GetBucketLocation",
      "s3:ListAllMyBuckets",
    ]
    resources = ["*"]
  }

  statement {
    sid = "CloudWatchReadForMetrics"
    actions = [
      "cloudwatch:ListMetrics",
      "cloudwatch:GetMetricData",
      "cloudwatch:GetMetricStatistics"
    ]
    resources = ["*"]
  }

  statement {
    sid = "AllowChangeOwnPassword"
    actions = ["iam:ChangePassword"]
    resources = [aws_iam_user.read_only_user.arn]
  }
}

data "aws_iam_policy_document" "lambda_app_policy" {
  statement {
    sid = "S3UploadAccess"
    actions = [
      "s3:GetObject",
      "s3:PutObject",
      "s3:ListBucket"
    ]
    resources = ["*"]
  }

  statement {
    sid = "SESAccess"
    actions = [
      "ses:SendEmail",
      "ses:SendRawEmail"
    ]
    resources = ["*"]
  }

  statement {
    sid = "SQSAccess"
    actions = [ 
      "sqs:SendMessage",
      "sqs:ReceiveMessage",
      "sqs:DeleteMessage",
      "sqs:GetQueueAttributes"
    ]
    resources = ["*"]
  }

   statement {
    sid = "AllowChangeOwnPassword"
    actions = ["iam:ChangePassword"]
    resources = [aws_iam_user.lambda_app.arn]
  }
}

