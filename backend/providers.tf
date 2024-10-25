terraform {
  required_version = "~> 1.9.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.73.0"
    }
  }
}

provider "aws" {
  region = "us-east-1"  # Primary region for backend storage

  default_tags {
    tags = {
      Environment = "shared"
      Project     = var.project_name
      ManagedBy   = "terraform"
      Component   = "backend"
    }
  }
}
