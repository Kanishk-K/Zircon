# This file sets up the dynamodb table to store metadata on jobs and users.
# -> Jobs Table
# -> Users Table

# CREATES a DynamoDB table to store metadata on jobs
resource "aws_dynamodb_table" "jobs-table" {
  name           = "Jobs"
  billing_mode   = "PROVISIONED"
  read_capacity  = 20
  write_capacity = 20
  hash_key       = "entryID"
  attribute {
    name = "entryID"
    type = "S"
  }
  tags = {
    Name        = "lecture-analyzer-jobs-table"
    Environment = "prod"
  }
}

# CREATES a DynamoDB table to store metadata on users
resource "aws_dynamodb_table" "users-table" {
  name           = "Users"
  billing_mode   = "PROVISIONED"
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "userID"
  attribute {
    name = "userID"
    type = "S"
  }
  tags = {
    Name        = "lecture-analyzer-users-table"
    Environment = "prod"
  }
}
