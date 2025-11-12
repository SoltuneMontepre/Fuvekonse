# Doppler Secrets Data Source
# Fetches all secrets from the configured Doppler project and config
# The Doppler token should be scoped to the appropriate project/config

data "doppler_secrets" "this" {}

# Local values for easier secret access throughout the configuration
locals {
  secrets = data.doppler_secrets.this.map

  # Database secrets
  db_host     = local.secrets.DB_HOST
  db_port     = local.secrets.DB_PORT
  db_user     = local.secrets.DB_USER
  db_password = local.secrets.DB_PASSWORD
  db_name     = local.secrets.DB_NAME
  db_sslmode  = local.secrets.DB_SSLMODE

  # Redis secrets
  redis_host = local.secrets.REDIS_HOST
  redis_port = local.secrets.REDIS_PORT
  redis_url  = local.secrets.REDIS_URL

  # Application secrets
  jwt_secret                      = local.secrets.JWT_SECRET
  jwt_access_token_expiry_minutes = local.secrets.JWT_ACCESS_TOKEN_EXPIRY_MINUTES
  jwt_refresh_token_expiry_days   = local.secrets.JWT_REFRESH_TOKEN_EXPIRY_DAYS
  login_max_fail                  = local.secrets.LOGIN_MAX_FAIL
  login_fail_block_minutes        = local.secrets.LOGIN_FAIL_BLOCK_MINUTES
  frontend_url                    = local.secrets.FRONTEND_URL
  
  # SES secrets
  ses_sender_email = local.secrets.SES_SENDER_EMAIL
}
