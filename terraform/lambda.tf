# This file creates the Lambda functions that the api gateway will use
# -> Local Variable Setup
# -> Zip Golang Binary

# -> Login Lambda
# -> Callback Lambda
# -> Auth Lambda
# -> Job Lambda
# -> Subtitle Lambda
# -> Queue Lambda

locals {
  zip_path = "${path.module}/../backend/bin"
}

resource "aws_lambda_function" "pre_signup_lambda" {
  function_name    = "zircon-pre-signup-lambda"
  role             = aws_iam_role.lambda-role.arn
  runtime          = "provided.al2023"
  handler          = "bootstrap"
  filename         = "${local.zip_path}/PreSignUp.zip"
  source_code_hash = filebase64sha256("${local.zip_path}/PreSignUp.zip")
  memory_size      = 128
  environment {
    variables = {
      ORGANIZATION = "@umn.edu"
    }
  }
}

resource "aws_lambda_function" "post_signup_lambda" {
  function_name    = "zircon-post-signup-lambda"
  role             = aws_iam_role.post_signup_lambda_role.arn
  runtime          = "provided.al2023"
  handler          = "bootstrap"
  filename         = "${local.zip_path}/PostSignUp.zip"
  source_code_hash = filebase64sha256("${local.zip_path}/PostSignUp.zip")
  memory_size      = 128
}

resource "aws_lambda_function" "job-lambda" {
  function_name    = "zircon-job-lambda"
  role             = aws_iam_role.submit-job-role.arn
  runtime          = "provided.al2023"
  handler          = "bootstrap"
  filename         = "${local.zip_path}/Job.zip"
  source_code_hash = filebase64sha256("${local.zip_path}/Job.zip")
  memory_size      = 128
  timeout          = 29
  environment {
    variables = {
      OPENAI_API_KEY     = var.OPENAI_API_KEY
      KALTURA_PARTNER_ID = var.KALTURA_PARTNER_ID
    }
  }
}

resource "aws_lambda_function" "subtitle-lambda" {
  function_name    = "zircon-subtitle-lambda"
  role             = aws_iam_role.tts-role.arn
  runtime          = "provided.al2023"
  handler          = "bootstrap"
  filename         = "${local.zip_path}/Subtitles.zip"
  source_code_hash = filebase64sha256("${local.zip_path}/Subtitles.zip")
  memory_size      = 128
  timeout          = 30
  environment {
    variables = {
      LEMONFOX_API_KEY = var.LEMONFOX_API_KEY
    }
  }
}

resource "aws_lambda_function" "queue-lambda" {
  function_name    = "zircon-queue-lambda"
  role             = aws_iam_role.queue-lambda.arn
  runtime          = "provided.al2023"
  handler          = "bootstrap"
  filename         = "${local.zip_path}/Queue.zip"
  source_code_hash = filebase64sha256("${local.zip_path}/Queue.zip")
  memory_size      = 128
  vpc_config {
    security_group_ids = [aws_security_group.lambda-elasticache-sg.id]
    subnet_ids         = aws_subnet.public-subnets[*].id
  }
  environment {
    variables = {
      REDIS_URL = "${aws_elasticache_replication_group.task-queue.primary_endpoint_address}:6379"
    }
  }
}

resource "aws_lambda_function" "exists_lambda" {
  function_name    = "zircon-exists-lambda"
  role             = aws_iam_role.exists_lambda_role.arn
  runtime          = "provided.al2023"
  handler          = "bootstrap"
  filename         = "${local.zip_path}/Exists.zip"
  source_code_hash = filebase64sha256("${local.zip_path}/Exists.zip")
  memory_size      = 128
}

resource "aws_lambda_function" "health_lambda" {
  function_name    = "zircon-health-lambda"
  role             = aws_iam_role.health_lambda.arn
  runtime          = "provided.al2023"
  handler          = "bootstrap"
  filename         = "${local.zip_path}/Health.zip"
  source_code_hash = filebase64sha256("${local.zip_path}/Health.zip")
  memory_size      = 128
  vpc_config {
    security_group_ids = [aws_security_group.lambda-elasticache-sg.id]
    subnet_ids         = aws_subnet.public-subnets[*].id
  }
  environment {
    variables = {
      REDIS_URL = "${aws_elasticache_replication_group.task-queue.primary_endpoint_address}:6379"
    }
  }
}

resource "aws_lambda_function" "ttl_video" {
  function_name    = "zircon-ttl-video-lambda"
  role             = aws_iam_role.ttl-role.arn
  runtime          = "provided.al2023"
  handler          = "bootstrap"
  filename         = "${local.zip_path}/TTLVideo.zip"
  source_code_hash = filebase64sha256("${local.zip_path}/TTLVideo.zip")
  memory_size      = 128
}
