terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
    doppler = {
      source  = "DopplerHQ/doppler"
      version = "~> 1.0"
    }
  }

  backend "s3" {
    bucket         = "fuvekon-terraform-state"
    key            = "fuvekon/terraform.tfstate"
    region         = "ap-southeast-1"
    encrypt        = true
    use_lockfile   = true
  }
}

provider "aws" {
  region = var.aws_region
}

provider "doppler" {
  doppler_token = var.doppler_token
}
