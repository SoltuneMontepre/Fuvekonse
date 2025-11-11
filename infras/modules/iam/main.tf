resource "aws_iam_user" "read_only_user" {
  name = var.iam_bucket_access_username

  tags = {
    Name        = var.project_name
    Environment = "Production"
  }
}

resource "aws_iam_user_login_profile" "read_only_user_login" {
  user    = aws_iam_user.read_only_user.name
  password_reset_required = true

  depends_on = [aws_iam_user.read_only_user]
}