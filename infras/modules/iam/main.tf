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

resource "aws_iam_user_policy" "read_only_user_inline" {
  name = "read-only-user-inline-policy"
  user = aws_iam_user.read_only_user.name
  policy = data.aws_iam_policy_document.read_only_user_policy.json
}