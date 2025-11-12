resource "aws_s3_bucket_policy" "this" {
  bucket = aws_s3_bucket.this.id
  policy = data.aws_iam_policy_document.this.json
}

data "aws_iam_policy_document" "this" {
  statement {
    sid = "PublicReadGetObject"
    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
    actions   = ["s3:GetObject"]
    resources = ["${aws_s3_bucket.this.arn}/*"]
  }

  statement {
    sid = "SpecificPrincipalsRead"
    actions = ["s3:GetObject", "s3:ListBucket"]
    resources = [aws_s3_bucket.this.arn, "${aws_s3_bucket.this.arn}/*"]

    dynamic "principals" {
      for_each = length(var.read_only_principal_arns) > 0 ? [1] : []
      content {
        type        = "AWS"
        identifiers = var.read_only_principal_arns
      }
    }
  }
}
