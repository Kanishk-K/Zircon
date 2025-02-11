# This file sets up the various IAM roles needed by the application.
# -> General Role Setup
# -> Lambda Role Setup
# -> ECS Consumer Task Role (Applied to EC2 consumer instances)
# -> ECS Consumer Task Execution Role (Applied to the ECS consumer service)
# -> S3 CloudFront Origin Access Control Policy

# DEFINE the trust policy for the ECS Task Role
data "aws_iam_policy_document" "ecs-task-trust-policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com", "ecs-tasks.amazonaws.com"]
    }
  }
}

# DEFINE the trust policy for the ECS Task Execution Role
data "aws_iam_policy_document" "ecs-task-execution-trust-policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

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

#
# Consumer Setup
#

# CREATE the ECS Consumer Task Role
resource "aws_iam_role" "ecs-consumer-task-role" {
  name               = "lecture-analyzer-ecs-consumer-task-role"
  assume_role_policy = data.aws_iam_policy_document.ecs-task-trust-policy.json
}

# ATTACH the required policy to the ECS Consumer Task Role
resource "aws_iam_role_policy_attachment" "ecs-consumer-task-role-policy" {
  role       = aws_iam_role.ecs-consumer-task-role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

# CREATE a profile for the ECS Consumer Task Role
resource "aws_iam_instance_profile" "ecs-consumer-task-profile" {
  name = "lecture-analyzer-ecs-consumer-task-profile"
  role = aws_iam_role.ecs-consumer-task-role.name
}

# DEFINE the dynamodb access policy for the ECS Consumer Task Role
data "aws_iam_policy_document" "ecs-consumer-task-dynamodb" {
  statement {
    actions = ["dynamodb:GetItem", "dynamodb:UpdateItem"]
    resources = [
      aws_dynamodb_table.jobs-table.arn,
    ]
  }
}

# CREATE the dynamodb access policy for the ECS Consumer Task Role
resource "aws_iam_policy" "ecs-consumer-task-dynamodb" {
  name        = "lecture-analyzer-ecs-consumer-task-dynamodb"
  description = "Allows the task to access DynamoDB"
  policy      = data.aws_iam_policy_document.ecs-consumer-task-dynamodb.json
}

# ATTACH the dynamodb access policy to the ECS Consumer Task Role
resource "aws_iam_role_policy_attachment" "ecs-consumer-task-policy" {
  role       = aws_iam_role.ecs-consumer-task-role.name
  policy_arn = aws_iam_policy.ecs-consumer-task-dynamodb.arn
}

# DEFINE the s3 access policy for the ECS Consumer Task Role
data "aws_iam_policy_document" "ecs-consumer-task-s3" {
  statement {
    effect  = "Allow"
    actions = ["s3:GetObject", "s3:PutObject", "s3:DeleteObject"]
    resources = [
      "${aws_s3_bucket.s3_bucket.arn}/assets/*"
    ]
  }
  statement {
    effect  = "Allow"
    actions = ["s3:GetObject"]
    resources = [
      "${aws_s3_bucket.s3_bucket.arn}/background/*"
    ]
  }
  statement {
    effect  = "Allow"
    actions = ["s3:ListBucket"]
    resources = [
      aws_s3_bucket.s3_bucket.arn
    ]
  }
}

# CREATE the s3 access policy for the ECS Consumer Task Role
resource "aws_iam_policy" "ecs-consumer-task-s3" {
  name        = "lecture-analyzer-ecs-consumer-task-s3"
  description = "Allows the task to access S3"
  policy      = data.aws_iam_policy_document.ecs-consumer-task-s3.json
}

# ATTACH the s3 access policy to the ECS Consumer Task Role
resource "aws_iam_role_policy_attachment" "ecs-consumer-task-s3-policy" {
  role       = aws_iam_role.ecs-consumer-task-role.name
  policy_arn = aws_iam_policy.ecs-consumer-task-s3.arn
}

# DEFINE the polly access policy for the ECS Consumer Task Role
data "aws_iam_policy_document" "ecs-consumer-task-polly" {
  statement {
    actions   = ["polly:StartSpeechSynthesisTask", "polly:GetSpeechSynthesisTask"]
    resources = ["*"]
  }
}

# CREATE the polly access policy for the ECS Consumer Task Role
resource "aws_iam_policy" "ecs-consumer-task-polly" {
  name        = "lecture-analyzer-ecs-consumer-task-polly"
  description = "Allows the task to access Polly"
  policy      = data.aws_iam_policy_document.ecs-consumer-task-polly.json
}

# ATTACH the polly access policy to the ECS Consumer Task Role
resource "aws_iam_role_policy_attachment" "ecs-consumer-task-polly-policy" {
  role       = aws_iam_role.ecs-consumer-task-role.name
  policy_arn = aws_iam_policy.ecs-consumer-task-polly.arn
}

# CREATE the ECS Consumer Task Execution Role
resource "aws_iam_role" "ecs-consumer-task-execution-role" {
  name               = "lecture-analyzer-ecs-consumer-task-execution-role"
  assume_role_policy = data.aws_iam_policy_document.ecs-task-execution-trust-policy.json
}

# ATTACH the role to the required ECS Consumer Task Execution Role policy
resource "aws_iam_role_policy_attachment" "ecs-consumer-task-execution-role-policy" {
  role       = aws_iam_role.ecs-consumer-task-execution-role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# DEFINE the secrets access policy for the ECS Consumer Task Execution Role
data "aws_iam_policy_document" "ecs-consumer-task-execution-secrets" {
  statement {
    actions = ["ssm:GetParameter", "ssm:GetParameters"]
    resources = [
      aws_ssm_parameter.OPENAI_API_KEY.arn
    ]
  }
}

# CREATE the ECS Consumer Task Execution Role resource access policy
resource "aws_iam_policy" "ecs-consumer-task-execution-secrets" {
  name        = "lecture-analyzer-ecs-consumer-task-execution-secrets"
  description = "Allows the Consumer ECS controller to access resources"
  policy      = data.aws_iam_policy_document.ecs-consumer-task-execution-secrets.json
}

# ATTACH the secrets access policy to the ECS Consumer Task Execution Role
resource "aws_iam_role_policy_attachment" "ecs-consumer-task-execution-policy" {
  role       = aws_iam_role.ecs-consumer-task-execution-role.name
  policy_arn = aws_iam_policy.ecs-consumer-task-execution-secrets.arn
}

# DEFINE the s3 access policy for CloudFront
data "aws_iam_policy_document" "read_content" {
  # Allow access to the content in the bucket
  statement {
    principals {
      type        = "Service"
      identifiers = ["cloudfront.amazonaws.com"]
    }
    actions   = ["s3:GetObject"]
    resources = ["${aws_s3_bucket.s3_bucket.arn}/assets/*"]
    condition {
      test     = "StringEquals"
      variable = "AWS:SourceArn"
      values   = [aws_cloudfront_distribution.web_routing.arn]
    }
  }
}
