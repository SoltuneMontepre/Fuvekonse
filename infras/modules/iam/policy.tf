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

data "aws_iam_policy_document" "s3_upload_user_inline_policy" {
  statement {
    sid = "S3ObjectAccess"
    actions = [
      "s3:PutObject",
      "s3:PutObjectAcl",
      "s3:GetObject",
      "s3:DeleteObject",
      "s3:ListBucketMultipartUploads",
      "s3:AbortMultipartUpload",
      "s3:ListMultipartUploadParts"
    ]
    resources = [
      "arn:aws:s3:::${var.s3_bucket_name}/*"
    ]
  }

  statement {
    sid = "S3BucketAccess"
    actions = [
      "s3:ListBucket",
      "s3:GetBucketLocation"
    ]
    resources = [
      "arn:aws:s3:::${var.s3_bucket_name}"
    ]
  }
}