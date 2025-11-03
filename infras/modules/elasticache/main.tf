# ElastiCache Subnet Group
resource "aws_elasticache_subnet_group" "this" {
  name       = "${var.name_prefix}-cache-subnet-group"
  subnet_ids = var.subnet_ids

  tags = merge(
    var.tags,
    {
      Name = "${var.name_prefix}-cache-subnet-group"
    }
  )
}

# ElastiCache Replication Group
resource "aws_elasticache_replication_group" "this" {
  replication_group_id       = var.replication_group_id
  description                = var.description
  engine                     = "redis"
  engine_version             = var.engine_version
  node_type                  = var.node_type
  num_cache_clusters         = var.num_cache_clusters
  parameter_group_name       = var.parameter_group_name
  port                       = var.port
  subnet_group_name          = aws_elasticache_subnet_group.this.name
  security_group_ids         = var.security_group_ids
  automatic_failover_enabled = var.automatic_failover_enabled
  multi_az_enabled           = var.multi_az_enabled
  at_rest_encryption_enabled = var.at_rest_encryption_enabled
  transit_encryption_enabled = var.transit_encryption_enabled
  auth_token                 = var.auth_token_enabled ? var.auth_token : null
  snapshot_retention_limit   = var.snapshot_retention_limit
  snapshot_window            = var.snapshot_window
  maintenance_window         = var.maintenance_window
  notification_topic_arn     = var.notification_topic_arn
  apply_immediately          = var.apply_immediately

  log_delivery_configuration {
    destination      = var.slow_log_destination
    destination_type = var.slow_log_destination_type
    log_format       = var.log_format
    log_type         = "slow-log"
  }

  log_delivery_configuration {
    destination      = var.engine_log_destination
    destination_type = var.engine_log_destination_type
    log_format       = var.log_format
    log_type         = "engine-log"
  }

  tags = merge(
    var.tags,
    {
      Name = var.replication_group_id
    }
  )
}
