# This file sets up the dynamodb table to store metadata on jobs and users.
# -> Jobs Table
# -> Users Table

# CREATES a DynamoDB table to store metadata on jobs
resource "aws_dynamodb_table" "jobs-table" {
  name             = "Jobs"
  billing_mode     = "PROVISIONED"
  read_capacity    = 20
  write_capacity   = 20
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

resource "aws_lambda_event_source_mapping" "invoke-video" {
  event_source_arn       = aws_dynamodb_table.jobs-table.stream_arn
  function_name          = aws_lambda_function.queue-lambda.arn
  starting_position      = "LATEST"
  batch_size             = 1
  enabled                = true
  maximum_retry_attempts = 0
  filter_criteria {
    filter {
      pattern = replace(replace(jsonencode({
        eventName = ["MODIFY"]
        dynamodb = {
          NewImage = {
            # Check that numVideos is not equal to 0
            videosAvailable = {
              SS = [
                { exists = true }
              ]
            }
          }
        }
      }), "\\u003c", "<"), "\\u003e", ">")
    }
  }
}
resource "aws_lambda_function_event_invoke_config" "video-noretry" {
  function_name          = aws_lambda_function.queue-lambda.function_name
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
