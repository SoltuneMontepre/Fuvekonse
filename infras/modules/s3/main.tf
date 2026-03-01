resource "aws_s3_bucket" "this" {
  bucket = var.bucket_name


  tags = {
    Name        = var.project_name
    Environment = "Production"
  }
}

