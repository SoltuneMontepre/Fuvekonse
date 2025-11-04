# Local values
locals {
  common_tags = merge(
    var.tags,
    {
      Project     = var.project_name
      Environment = var.environment
      ManagedBy   = "Terraform"
    }
  )
}

################################################################################
# VPC Module
################################################################################

module "vpc" {
  source = "./modules/vpc"

  name_prefix          = var.project_name
  vpc_cidr             = var.vpc_cidr
  availability_zones   = var.availability_zones
  public_subnet_cidrs  = var.public_subnet_cidrs
  private_subnet_cidrs = var.private_subnet_cidrs
  single_nat_gateway   = true

  tags = local.common_tags
}

################################################################################
# Security Groups
################################################################################

module "lambda_sg" {
  source = "./modules/security-group"

  name        = "$${var.project_name}-lambda-sg"
  description = "Security group for Lambda functions"
  vpc_id      = module.vpc.vpc_id

  egress_rules = [
    {
      from_port   = 0
      to_port     = 0
      protocol    = "-1"
      cidr_blocks = ["0.0.0.0/0"]
      description = "Allow all outbound traffic"
    }
  ]

  tags = local.common_tags
}

module "rds_sg" {
  source = "./modules/security-group"

  name        = "$${var.project_name}-rds-sg"
  description = "Security group for RDS database"
  vpc_id      = module.vpc.vpc_id

  ingress_rules = [
    {
      from_port       = 5432
      to_port         = 5432
      protocol        = "tcp"
      security_groups = [module.lambda_sg.security_group_id]
      description     = "PostgreSQL from Lambda"
    }
  ]

  egress_rules = [
    {
      from_port   = 0
      to_port     = 0
      protocol    = "-1"
      cidr_blocks = ["0.0.0.0/0"]
      description = "Allow all outbound traffic"
    }
  ]

  tags = local.common_tags
}

module "elasticache_sg" {
  source = "./modules/security-group"

  name        = "$${var.project_name}-elasticache-sg"
  description = "Security group for ElastiCache"
  vpc_id      = module.vpc.vpc_id

  ingress_rules = [
    {
      from_port       = 6379
      to_port         = 6379
      protocol        = "tcp"
      security_groups = [module.lambda_sg.security_group_id]
      description     = "Redis from Lambda"
    }
  ]

  egress_rules = [
    {
      from_port   = 0
      to_port     = 0
      protocol    = "-1"
      cidr_blocks = ["0.0.0.0/0"]
      description = "Allow all outbound traffic"
    }
  ]

  tags = local.common_tags
}

################################################################################
# RDS Module
################################################################################

module "rds" {
  source = "./modules/rds"

  name_prefix            = var.project_name
  identifier             = "$${var.project_name}-db"
  engine                 = "postgres"
  engine_version         = var.db_engine_version
  instance_class         = var.db_instance_class
  allocated_storage      = var.db_allocated_storage
  db_name                = var.db_name
  username               = var.db_username
  password               = var.db_password
  subnet_ids             = module.vpc.private_subnet_ids
  vpc_security_group_ids = [module.rds_sg.security_group_id]
  multi_az               = true
  skip_final_snapshot    = false

  tags = local.common_tags
}

################################################################################
# ElastiCache Module
################################################################################

module "elasticache" {
  source = "./modules/elasticache"

  name_prefix          = var.project_name
  replication_group_id = "$${var.project_name}-cache"
  description          = "Redis cache for $${var.project_name}"
  engine_version       = var.redis_engine_version
  node_type            = var.redis_node_type
  num_cache_clusters   = var.redis_num_cache_nodes
  subnet_ids           = module.vpc.private_subnet_ids
  security_group_ids   = [module.elasticache_sg.security_group_id]

  tags = local.common_tags
}

################################################################################
# S3 Module
################################################################################

module "s3" {
  source = "./modules/s3"

  bucket_name         = var.bucket_name
  versioning_enabled  = true
  encryption_enabled  = true
  block_public_access = true

  tags = local.common_tags
}

################################################################################
# SQS Module
################################################################################

module "sqs" {
  source = "./modules/sqs"

  name                       = "$${var.project_name}-message-queue"
  visibility_timeout_seconds = var.sqs_visibility_timeout
  create_dlq                 = true

  tags = local.common_tags
}

################################################################################
# IAM Module
################################################################################

module "iam" {
  source = "./modules/iam"

  project_name      = var.project_name
  s3_bucket_arns    = [module.s3.bucket_arn]
  sqs_queue_arns    = [module.sqs.queue_arn]
  enable_ses_access = true

  tags = local.common_tags
}

################################################################################
# Lambda Functions
################################################################################

module "lambda_general_service" {
  source = "./modules/lambda"

  function_name = "$${var.project_name}-general-service"
  filename      = "placeholder.zip"
  role_arn      = module.iam.lambda_execution_role_arn
  handler       = "main"
  runtime       = "provided.al2"
  timeout       = var.lambda_timeout
  memory_size   = var.lambda_memory_size

  vpc_config = {
    subnet_ids         = module.vpc.private_subnet_ids
    security_group_ids = [module.lambda_sg.security_group_id]
  }

  environment_variables = {
    DB_HOST     = module.rds.db_instance_address
    DB_PORT     = tostring(module.rds.db_instance_port)
    DB_NAME     = var.db_name
    DB_USERNAME = var.db_username
    DB_PASSWORD = var.db_password
    REDIS_HOST  = module.elasticache.primary_endpoint_address
    S3_BUCKET   = module.s3.bucket_id
  }

  tags = local.common_tags
}

module "lambda_ticket_service" {
  source = "./modules/lambda"

  function_name = "$${var.project_name}-ticket-service"
  filename      = "placeholder.zip"
  role_arn      = module.iam.lambda_execution_role_arn
  handler       = "main"
  runtime       = "provided.al2"
  timeout       = var.lambda_timeout
  memory_size   = var.lambda_memory_size

  vpc_config = {
    subnet_ids         = module.vpc.private_subnet_ids
    security_group_ids = [module.lambda_sg.security_group_id]
  }

  environment_variables = {
    DB_HOST       = module.rds.db_instance_address
    DB_PORT       = tostring(module.rds.db_instance_port)
    DB_NAME       = var.db_name
    DB_USERNAME   = var.db_username
    DB_PASSWORD   = var.db_password
    REDIS_HOST    = module.elasticache.primary_endpoint_address
    SQS_QUEUE_URL = module.sqs.queue_url
  }

  tags = local.common_tags
}

module "lambda_worker_node" {
  source = "./modules/lambda"

  function_name = "$${var.project_name}-worker-node"
  filename      = "placeholder.zip"
  role_arn      = module.iam.lambda_execution_role_arn
  handler       = "main"
  runtime       = "provided.al2"
  timeout       = 300
  memory_size   = 512

  vpc_config = {
    subnet_ids         = module.vpc.private_subnet_ids
    security_group_ids = [module.lambda_sg.security_group_id]
  }

  environment_variables = {
    SQS_QUEUE_URL  = module.sqs.queue_url
    SES_FROM_EMAIL = "noreply@$${var.domain_name}"
  }

  event_source_arn        = module.sqs.queue_arn
  event_source_batch_size = 10

  tags = local.common_tags
}

################################################################################
# SES (Simple Email Service)
################################################################################

resource "aws_ses_domain_identity" "main" {
  domain = var.domain_name
}
