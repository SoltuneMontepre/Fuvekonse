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

# Ticket queue (general-service + sqs-worker). Required for ticket write flow in production.
# general_service_url = "https://<api-gateway-id>.execute-api.ap-southeast-1.amazonaws.com/api/general"  # Use: terraform output general_service_url
# internal_api_key   = "<secret>"  # Same value for both Lambdas; store in Doppler or secrets manager (sensitive)
