output "db_instance_id" {
  description = "ID of the DB instance"
  value       = aws_db_instance.this.id
}

output "db_instance_arn" {
  description = "ARN of the DB instance"
  value       = aws_db_instance.this.arn
}

output "db_instance_endpoint" {
  description = "Connection endpoint"
  value       = aws_db_instance.this.endpoint
  sensitive   = true
}

output "db_instance_address" {
  description = "Address of the DB instance"
  value       = aws_db_instance.this.address
  sensitive   = true
}

output "db_instance_port" {
  description = "Port of the DB instance"
  value       = aws_db_instance.this.port
}

output "db_subnet_group_id" {
  description = "ID of the DB subnet group"
  value       = aws_db_subnet_group.this.id
}

output "db_name" {
  description = "Name of the database"
  value       = aws_db_instance.this.db_name
}
