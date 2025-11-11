module "s3" {
  source       = "./modules/s3"
  project_name = var.project_name
  bucket_name  = var.bucket_name
  bucket_acl   = var.bucket_acl
  read_only_principal_arns = [module.iam.read_only_user_arn]
}

module "iam" {
  source                 = "./modules/iam"
  project_name           = var.project_name
  iam_bucket_access_username = var.iam_bucket_access_username
}