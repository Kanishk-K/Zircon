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

resource "aws_lambda_function" "login-lambda" {
  function_name    = "zircon-login-lambda"
  role             = aws_iam_role.lambda-role.arn
  runtime          = "provided.al2023"
  handler          = "bootstrap"
  filename         = "${local.zip_path}/Login.zip"
  source_code_hash = filebase64sha256("${local.zip_path}/Login.zip")
  memory_size      = 128
  environment {
    variables = {
      GOOGLE_CLIENT_ID     = var.GOOGLE_CLIENT_ID
      GOOGLE_CLIENT_SECRET = var.GOOGLE_CLIENT_SECRET
      HOST                 = "https://${var.DOMAIN}"
    }
  }
}

resource "aws_lambda_function" "callback-lambda" {
  function_name    = "zircon-callback-lambda"
  role             = aws_iam_role.lambda-callback-role.arn
  runtime          = "provided.al2023"
  handler          = "bootstrap"
  filename         = "${local.zip_path}/Callback.zip"
  source_code_hash = filebase64sha256("${local.zip_path}/Callback.zip")
  memory_size      = 128
  environment {
    variables = {
      GOOGLE_CLIENT_ID     = var.GOOGLE_CLIENT_ID
      GOOGLE_CLIENT_SECRET = var.GOOGLE_CLIENT_SECRET
      JWT_PRIVATE          = var.JWT_PRIVATE
      HOST                 = "https://${var.DOMAIN}"
    }
  }
}

resource "aws_lambda_function" "auth-lambda" {
  function_name    = "zircon-auth-lambda"
  role             = aws_iam_role.lambda-authorizer-role.arn
  runtime          = "provided.al2023"
  handler          = "bootstrap"
  filename         = "${local.zip_path}/Auth.zip"
  source_code_hash = filebase64sha256("${local.zip_path}/Auth.zip")
  memory_size      = 128
  environment {
    variables = {
      JWT_PRIVATE = var.JWT_PRIVATE
    }
  }
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
      OPENAI_API_KEY = var.OPENAI_API_KEY
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
