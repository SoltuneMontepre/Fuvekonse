data "doppler_secrets" "this" {
  project = var.doppler_project
  config  = var.doppler_config
}

locals {
  secrets = data.doppler_secrets.this.map

  db_host     = local.secrets.DB_HOST
  db_port     = local.secrets.DB_PORT
  db_user     = local.secrets.DB_USER
  db_password = local.secrets.DB_PASSWORD
  db_name     = local.secrets.DB_NAME
  db_sslmode  = local.secrets.DB_SSLMODE

  redis_host = local.secrets.REDIS_HOST
  redis_port = local.secrets.REDIS_PORT
  redis_url  = local.secrets.REDIS_URL

  jwt_secret                      = local.secrets.JWT_SECRET
  jwt_access_token_expiry_minutes = local.secrets.JWT_ACCESS_TOKEN_EXPIRY_MINUTES
  jwt_refresh_token_expiry_days   = local.secrets.JWT_REFRESH_TOKEN_EXPIRY_DAYS
  login_max_fail                  = local.secrets.LOGIN_MAX_FAIL
  login_fail_block_minutes        = local.secrets.LOGIN_FAIL_BLOCK_MINUTES
  frontend_url                    = local.secrets.FRONTEND_URL
  
  ses_sender_email = local.secrets.SES_SENDER_EMAIL
}
