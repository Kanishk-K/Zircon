# This file creates the Lambda functions that the api gateway will use
# -> Local Variable Setup
# -> Zip Golang Binary

# -> Auth Lambda
# -> Callback Lambda
# -> Status Lambda
# -> Process Lambda

locals {
  zip_path = "${path.module}/../backend/bin"
}

resource "aws_lambda_function" "auth-lambda" {
  function_name    = "zircon-auth-lambda"
  role             = aws_iam_role.lambda-role.arn
  runtime          = "provided.al2023"
  handler          = "auth"
  filename         = "${local.zip_path}/Auth.zip"
  source_code_hash = filebase64sha256("${local.zip_path}/Auth.zip")
  memory_size      = 128
}


