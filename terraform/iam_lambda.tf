# DEFINE the trust policy for the Lambda Role
data "aws_iam_policy_document" "lambda-trust-policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "lambda-role" {
  name               = "lambda-role"
  assume_role_policy = data.aws_iam_policy_document.lambda-trust-policy.json
}

data "aws_iam_policy_document" "innerVPC-description" {
  statement {
    actions   = ["ec2:CreateNetworkInterface", "ec2:DescribeNetworkInterfaces", "ec2:DeleteNetworkInterface"]
    resources = ["*"]
  }
}

resource "aws_iam_policy" "innerVPC-policy" {
  name        = "innerVPC-policy"
  description = "Allows for lambdas to create themselves in a VPC"
  policy      = data.aws_iam_policy_document.innerVPC-description.json
}

resource "aws_iam_policy_attachment" "innerVPC-lambda-policy" {
  name       = "innerVPC-lambda-policy"
  roles      = [aws_iam_role.queue-lambda.name, aws_iam_role.health_lambda.name]
  policy_arn = aws_iam_policy.innerVPC-policy.arn
}

# Lambda logging policy attachment
resource "aws_iam_policy_attachment" "submit-cloudwatch" {
  name = "submit-cloudwatch"
  roles = [
    aws_iam_role.submit-job-role.name,
    aws_iam_role.tts-role.name,
    aws_iam_role.queue-lambda.name,
    aws_iam_role.exists_lambda_role.name,
    aws_iam_role.health_lambda.name,
    aws_iam_role.ttl-role.name,
  ]
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}
