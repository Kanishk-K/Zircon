# This file sets up the dynamodb table to store metadata on jobs and users.
# -> Jobs Table
# -> Video Request Table
# -> Users Table

# CREATES a DynamoDB table to store metadata on jobs
resource "aws_dynamodb_table" "jobs-table" {
  name             = "Jobs"
  billing_mode     = "PROVISIONED"
  read_capacity    = 15
  write_capacity   = 15
  hash_key         = "entryID"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"
  attribute {
    name = "entryID"
    type = "S"
  }
  tags = {
    Name        = "lecture-analyzer-jobs-table"
    Environment = "prod"
  }
  lifecycle {
    prevent_destroy = true
  }
}

resource "aws_lambda_event_source_mapping" "invoke-tts" {
  event_source_arn       = aws_dynamodb_table.jobs-table.stream_arn
  function_name          = aws_lambda_function.subtitle-lambda.arn
  starting_position      = "LATEST"
  batch_size             = 1
  enabled                = true
  maximum_retry_attempts = 0
  filter_criteria {
    filter {
      pattern = jsonencode({
        eventName = ["MODIFY"]
        dynamodb = {
          OldImage = {
            subtitlesGenerated = {
              BOOL = [false]
            }
          }
          NewImage = {
            subtitlesGenerated = {
              BOOL = [true]
            }
          }
        }
      })
    }
  }
}

resource "aws_lambda_function_event_invoke_config" "tts-noretry" {
  function_name          = aws_lambda_function.subtitle-lambda.function_name
  maximum_retry_attempts = 0
}

# CREATES a DynamoDB table to store metadata on video requests
resource "aws_dynamodb_table" "video_requests_table" {
  name             = "VideoRequests"
  billing_mode     = "PROVISIONED"
  read_capacity    = 5
  write_capacity   = 5
  hash_key         = "entryID"
  range_key        = "requestedVideo"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"
  ttl {
    attribute_name = "videoExpiry"
    enabled        = true
  }
  attribute {
    name = "entryID"
    type = "S"
  }
  attribute {
    name = "requestedVideo"
    type = "S"
  }
  tags = {
    Name        = "zircon-video-requests-table"
    Environment = "prod"
  }
  lifecycle {
    prevent_destroy = true
  }
}

resource "aws_lambda_event_source_mapping" "invoke-videogen" {
  event_source_arn       = aws_dynamodb_table.video_requests_table.stream_arn
  function_name          = aws_lambda_function.queue-lambda.arn
  starting_position      = "LATEST"
  batch_size             = 1
  enabled                = true
  maximum_retry_attempts = 0
  filter_criteria {
    filter {
      pattern = jsonencode({
        eventName = ["INSERT"]
      })
    }
  }
}

resource "aws_lambda_function_event_invoke_config" "videogen-noretry" {
  function_name          = aws_lambda_function.queue-lambda.function_name
  maximum_retry_attempts = 0
}

resource "aws_lambda_event_source_mapping" "invoke-ttl" {
  event_source_arn       = aws_dynamodb_table.video_requests_table.stream_arn
  function_name          = aws_lambda_function.ttl_video.arn
  starting_position      = "LATEST"
  batch_size             = 1
  enabled                = true
  maximum_retry_attempts = 0
  filter_criteria {
    filter {
      pattern = jsonencode({
        eventName = ["REMOVE"]
      })
    }
  }
}

resource "aws_lambda_function_event_invoke_config" "ttl-noretry" {
  function_name          = aws_lambda_function.ttl_video.function_name
  maximum_retry_attempts = 0
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
  lifecycle {
    prevent_destroy = true
  }
}
