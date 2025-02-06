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
  }
}


module "backend_bucket" {
  source      = "../../modules/s3_bucket"
  bucket_name = "zinsearch-backend-data"
}

output "backend_bucket_name" {
  value = module.backend_bucket.bucket_name
}

/*resource "aws_s3_bucket" "backend_bucket" {
  bucket = "zinsearch-backend-data"
}
output "backend_bucket_name" {
  value = aws_s3_bucket.backend_bucket.id
}*/
