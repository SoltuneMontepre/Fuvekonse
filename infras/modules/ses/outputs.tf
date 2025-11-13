output "sender_email" {
  description = "Verified SES sender email address"
  value       = aws_ses_email_identity.sender.email
}

output "sender_identity_arn" {
  description = "ARN of the SES email identity"
  value       = aws_ses_email_identity.sender.arn
}

output "configuration_set_name" {
  description = "Name of the SES configuration set"
  value       = aws_ses_configuration_set.main.name
}

output "configuration_set_arn" {
  description = "ARN of the SES configuration set"
  value       = aws_ses_configuration_set.main.arn
}
