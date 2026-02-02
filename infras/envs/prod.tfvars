aws_region  = "ap-southeast-1"
project_name = "fuvekon"

# Doppler Configuration
doppler_project = "fuvekon"
doppler_config  = "github-actions"

# S3 Configuration
bucket_name = "fuvekon-bucket"
bucket_acl  = "private"
s3_cors_allowed_origins = ["https://fuve.netlify.app", "http://localhost:3000", "http://localhost:5173"]

# IAM Configuration
iam_bucket_access_username = "bucket"
iam_lambda_app_role_name = "fuvekon-lambda-app-role"

# Lambda Deployment Packages
general_service_zip_path = "../services/general-service/bootstrap.zip"
rbac_service_zip_path = "../services/rbac-service/bootstrap.zip"
sqs_worker_zip_path = "../services/sqs-worker/bootstrap.zip"

# Application Configuration
gin_mode = "release"

# Internal API (general-service + sqs-worker). Uncomment and set for production.
# general_service_url = "https://<api-gateway-id>.execute-api.ap-southeast-1.amazonaws.com/api/general"
# internal_api_key   = "<secret>"
general_service_url = "https://riw96amgn7.execute-api.ap-southeast-1.amazonaws.com"
# internal_api_key set via TF_VAR_internal_api_key or -var to avoid committing secrets
