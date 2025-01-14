/*
  PRE-REQUISITES:
  - AWS CLI installed and configured with the necessary permissions.
  - Create a service-linked role for Elasticache.
*/

# General Terraform Setup
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

# Provision the S3 bucket for storing notes, summaries, and generated videos.
resource "aws_s3_bucket" "lecture-analyzer-content" {
  bucket        = "lecture-analyzer"
  force_destroy = true
  tags = {
    Name        = "lecture-analyzer"
    Environment = "prod"
  }
}

# Provision the DynamoDB table for storing the metadata of the generated content.
resource "aws_dynamodb_table" "lecture-analyzer-jobs" {
  name           = "Jobs"
  billing_mode   = "PROVISIONED"
  read_capacity  = 20
  write_capacity = 20
  hash_key       = "entryID"
  # Define the hashkey attribute
  attribute {
    name = "entryID"
    type = "S"
  }
  tags = {
    Name        = "lecture-analyzer"
    Environment = "prod"
  }
}

# Provision the DynamoDB table for storing user data.
resource "aws_dynamodb_table" "lecture-analyzer-users" {
  name           = "Users"
  billing_mode   = "PROVISIONED"
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "userID"
  # Define the hashkey attribute
  attribute {
    name = "userID"
    type = "S"
  }
  tags = {
    Name        = "lecture-analyzer"
    Environment = "prod"
  }
}

# Provision the VPC and subnets for the Lecture Analyzer application.
resource "aws_vpc" "lecture-analyzer-vpc" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name        = "lecture-analyzer"
    Environment = "prod"
  }
}

data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_subnet" "lecture-analyzer-private-subnet" {
  count             = length(data.aws_availability_zones.available.names)
  vpc_id            = aws_vpc.lecture-analyzer-vpc.id
  cidr_block        = cidrsubnet(aws_vpc.lecture-analyzer-vpc.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]
  tags = {
    Name        = "lecture-analyzer-private-subnet-${count.index}"
    Environment = "prod"
  }
}

resource "aws_elasticache_subnet_group" "lecture-analyzer-elasticate-subnet-group" {
  name       = "lecture-analyzer-elasticache-subnet-group"
  subnet_ids = aws_subnet.lecture-analyzer-private-subnet[*].id
}

# Provision the elasticache instance to handle job queueing.
resource "aws_elasticache_cluster" "lecture-analyzer-queue" {
  cluster_id           = "lecture-analyzer-queue"
  engine               = "redis"
  node_type            = "cache.t3.micro"
  num_cache_nodes      = 1
  parameter_group_name = "default.redis7"
  port                 = 6379
  subnet_group_name    = aws_elasticache_subnet_group.lecture-analyzer-elasticate-subnet-group.name
  tags = {
    Name        = "lecture-analyzer"
    Environment = "prod"
  }
}

