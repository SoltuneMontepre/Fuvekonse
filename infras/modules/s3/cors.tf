
resource "aws_s3_bucket_cors_configuration" "this" {
  bucket = aws_s3_bucket.this.id


  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "HEAD", "POST", "PUT"]
    allowed_origins = var.s3_cors_allowed_origins
    expose_headers  = ["ETag"]
    max_age_seconds = 3000
  }

  cors_rule {
    allowed_methods = ["GET", "HEAD", "POST", "PUT"]
    allowed_origins = var.s3_cors_allowed_origins
  }
}
