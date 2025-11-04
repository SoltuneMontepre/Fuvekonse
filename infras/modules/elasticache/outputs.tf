output "replication_group_id" {
  description = "ID of the ElastiCache replication group"
  value       = aws_elasticache_replication_group.this.id
}

output "replication_group_arn" {
  description = "ARN of the ElastiCache replication group"
  value       = aws_elasticache_replication_group.this.arn
}

output "primary_endpoint_address" {
  description = "Address of the primary endpoint"
  value       = aws_elasticache_replication_group.this.primary_endpoint_address
  sensitive   = true
}

output "reader_endpoint_address" {
  description = "Address of the reader endpoint"
  value       = aws_elasticache_replication_group.this.reader_endpoint_address
  sensitive   = true
}

output "configuration_endpoint_address" {
  description = "Address of the configuration endpoint (cluster mode enabled)"
  value       = try(aws_elasticache_replication_group.this.configuration_endpoint_address, null)
  sensitive   = true
}

output "member_clusters" {
  description = "List of member cluster IDs"
  value       = aws_elasticache_replication_group.this.member_clusters
}

output "subnet_group_name" {
  description = "Name of the subnet group"
  value       = aws_elasticache_subnet_group.this.name
}
