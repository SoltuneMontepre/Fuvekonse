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
ticket_service_zip_path = "../services/ticket-service/bootstrap.zip"
sqs_worker_zip_path = "../services/sqs-worker/bootstrap.zip"

# Application Configuration
gin_mode = "release"