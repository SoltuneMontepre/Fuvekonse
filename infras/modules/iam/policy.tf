
data "aws_iam_policy_document" "read_only_user_policy" {
  statement {
    sid = "S3Read"
    actions = [
      "s3:GetObject",
      "s3:ListBucket",
      "s3:GetBucketLocation",
      "s3:GetBucketMetricsConfiguration",
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

resource "aws_iam_user_policy" "read_only_user_inline" {
  name = "read-only-user-inline-policy"
  user = aws_iam_user.read_only_user.name
  policy = data.aws_iam_policy_document.read_only_user_policy.json
}