module "s3" {
  source                   = "./modules/s3"
  project_name             = var.project_name
  bucket_name              = var.bucket_name
  bucket_acl               = var.bucket_acl
  read_only_principal_arns = [module.iam.read_only_user_arn]
  s3_cors_allowed_origins  = var.s3_cors_allowed_origins
}

module "iam" {
  source                     = "./modules/iam"
  project_name               = var.project_name
  iam_bucket_access_username = var.iam_bucket_access_username
  iam_s3_upload_username     = var.iam_s3_upload_username
  s3_bucket_name             = module.s3.bucket_name
}

module "ses" {
  source       = "./modules/ses"
  project_name = var.project_name
  sender_email = local.ses_sender_email
}

module "sqs" {
  source       = "./modules/sqs"
  project_name = var.project_name
}

module "iam_role" {
  source                   = "./modules/iam_role"
  project_name             = var.project_name
  iam_lambda_app_role_name = var.iam_lambda_app_role_name
  s3_bucket_arn            = module.s3.bucket_arn
  ses_identity_arn         = module.ses.sender_identity_arn
  sqs_queue_arn            = module.sqs.queue_arn
  sqs_dlq_arn              = module.sqs.dead_letter_queue_arn
}

module "lambda" {
  source                          = "./modules/lambda"
  project_name                    = var.project_name
  lambda_role_arn                 = module.iam_role.lambda_app_role_arn
  general_service_zip_path        = var.general_service_zip_path
  rbac_service_zip_path           = var.rbac_service_zip_path
  sqs_worker_zip_path             = var.sqs_worker_zip_path
  db_host                         = local.db_host
  db_port                         = local.db_port
  db_user                         = local.db_user
  db_password                     = local.db_password
  db_name                         = local.db_name
  db_sslmode                      = local.db_sslmode
  redis_url                       = local.redis_url
  aws_region                      = var.aws_region
  jwt_secret                      = local.jwt_secret
  jwt_access_token_expiry_minutes = local.jwt_access_token_expiry_minutes
  jwt_refresh_token_expiry_days   = local.jwt_refresh_token_expiry_days
  login_max_fail                  = local.login_max_fail
  login_fail_block_minutes        = local.login_fail_block_minutes
  frontend_url                    = local.frontend_url
  gin_mode                        = var.gin_mode
  s3_bucket_name                  = module.s3.bucket_name
  ses_sender_email                = module.ses.sender_email
  sqs_queue_url                   = module.sqs.queue_url
  sqs_queue_arn                   = module.sqs.queue_arn
  cors_allowed_origins            = var.s3_cors_allowed_origins
  internal_api_key                = var.internal_api_key
  general_service_url             = var.general_service_url
}

module "networking" {
  source                        = "./modules/networking"
  project_name                  = var.project_name
  cors_allowed_origins          = var.s3_cors_allowed_origins
  general_service_function_name = module.lambda.general_service_function_name
  general_service_invoke_arn    = module.lambda.general_service_invoke_arn
  rbac_service_function_name    = module.lambda.rbac_service_function_name
  rbac_service_invoke_arn       = module.lambda.rbac_service_invoke_arn
}
