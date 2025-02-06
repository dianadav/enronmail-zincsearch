terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

provider "aws" {
  access_key                  = "test"
  secret_key                  = "test"
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  s3_use_path_style           = true
  endpoints {
    s3       = "http://localhost:4566"
    dynamodb = "http://localhost:4566"
    route53  = "http://localhost:4566"
  }
}

/*
resource "aws_s3_bucket" "frontend_bucket" {
  bucket = "enron-frontend-data"
}
output "frontend_bucket_name" {
  value = aws_s3_bucket.frontend_bucket.id
}
*/

module "frontend_bucket" {
  source      = "../../../modules/s3_bucket"
  bucket_name = "enron-frontend-data"
}

module "frontend_route53" {
  source      = "../../../modules/route53"
  domain_name = "frontend.local"
}

output "frontend_bucket_name" {
  value = module.frontend_bucket.bucket_name
}

output "frontend_domain_name" {
  value = module.frontend_route53.domain_name
}

