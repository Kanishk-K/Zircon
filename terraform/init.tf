/*
  PRE-REQUISITES:
  - AWS CLI installed and configured with the necessary permissions.
  - Create a service-linked role for Elasticache.
  - Store open_ai API key in AWS Parameter Store with the name /lecture-analyzer/openai-api-key.
*/

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.83.1"
    }
  }
}

provider "aws" {
  region = "us-east-1"
}
