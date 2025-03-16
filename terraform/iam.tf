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

# CREATE the Lambda Callback Role to allow writing to the users table
resource "aws_iam_role" "lambda-callback-role" {
  name               = "lambda-callback-role"
  assume_role_policy = data.aws_iam_policy_document.lambda-trust-policy.json
}

data "aws_iam_policy_document" "callback-dynamodb-description" {
  statement {
    actions = ["dynamodb:PutItem"]
    resources = [
      aws_dynamodb_table.users-table.arn,
    ]
  }
}

resource "aws_iam_policy" "callback-dynamodb" {
  name        = "callback-dynamodb"
  description = "Allows the callback lambda to write to the users table"
  policy      = data.aws_iam_policy_document.callback-dynamodb-description.json
}

resource "aws_iam_role_policy_attachment" "lambda-callback-dynamodb" {
  role       = aws_iam_role.lambda-callback-role.name
  policy_arn = aws_iam_policy.callback-dynamodb.arn
}

# CREATE the Lambda Authorizer Role for protecting the API Gateway
resource "aws_iam_role" "lambda-authorizer-role" {
  name               = "lambda-authorizer-role"
  assume_role_policy = data.aws_iam_policy_document.lambda-trust-policy.json
}

# CREATE the Lambda Submit Job Role
resource "aws_iam_role" "submit-job-role" {
  name               = "submit-job-role"
  assume_role_policy = data.aws_iam_policy_document.lambda-trust-policy.json
}

data "aws_iam_policy_document" "submit-dynamodb-description" {
  statement {
    actions = ["dynamodb:PutItem"]
    resources = [
      aws_dynamodb_table.jobs-table.arn,
      aws_dynamodb_table.video_requests_table.arn,
    ]
  }
  statement {
    actions = ["dynamodb:UpdateItem"]
    resources = [
      aws_dynamodb_table.jobs-table.arn,
      aws_dynamodb_table.users-table.arn,
    ]
  }
  statement {
    actions = ["dynamodb:DeleteItem"]
    resources = [
      aws_dynamodb_table.jobs-table.arn,
    ]
  }
}

resource "aws_iam_policy" "submit-dynamodb" {
  name        = "submit-dynamodb"
  description = "Allows the submit job lambda to write to the jobs and user tables"
  policy      = data.aws_iam_policy_document.submit-dynamodb-description.json
}

resource "aws_iam_role_policy_attachment" "lambda-submit-dynamodb" {
  role       = aws_iam_role.submit-job-role.name
  policy_arn = aws_iam_policy.submit-dynamodb.arn
}

data "aws_iam_policy_document" "submit-s3-description" {
  statement {
    actions = ["s3:PutObject"]
    resources = [
      "${aws_s3_bucket.s3_bucket.arn}/assets/*/Summary.txt",
      "${aws_s3_bucket.s3_bucket.arn}/assets/*/Notes.md",
    ]
  }
}

resource "aws_iam_policy" "submit-s3" {
  name        = "submit-s3"
  description = "Allows the submit job lambda to write summaries and notes to the S3 bucket"
  policy      = data.aws_iam_policy_document.submit-s3-description.json
}

resource "aws_iam_role_policy_attachment" "lambda-submit-s3" {
  role       = aws_iam_role.submit-job-role.name
  policy_arn = aws_iam_policy.submit-s3.arn
}

# CREATE the TTS Lambda role
resource "aws_iam_role" "tts-role" {
  name               = "tts-role"
  assume_role_policy = data.aws_iam_policy_document.lambda-trust-policy.json
}

data "aws_iam_policy_document" "jobs-stream-access-description" {
  statement {
    actions = ["dynamodb:GetRecords", "dynamodb:GetShardIterator", "dynamodb:DescribeStream", "dynamodb:ListStreams"]
    resources = [
      aws_dynamodb_table.jobs-table.stream_arn,
    ]
  }
}

resource "aws_iam_policy" "jobs-stream-access" {
  name        = "jobs-stream-access"
  description = "Allows the TTS lambda to access the jobs dynamodb stream"
  policy      = data.aws_iam_policy_document.jobs-stream-access-description.json
}

resource "aws_iam_policy_attachment" "jobs-lambda-stream-access" {
  name       = "jobs-lambda-stream-access"
  roles      = [aws_iam_role.tts-role.name]
  policy_arn = aws_iam_policy.jobs-stream-access.arn
}

data "aws_iam_policy_document" "videogen-stream-access-description" {
  statement {
    actions = ["dynamodb:GetRecords", "dynamodb:GetShardIterator", "dynamodb:DescribeStream", "dynamodb:ListStreams"]
    resources = [
      aws_dynamodb_table.video_requests_table.stream_arn,
    ]
  }
}

resource "aws_iam_policy" "videogen-stream-access" {
  name        = "videogen-stream-access"
  description = "Allows the TTS lambda to access the videogen dynamodb stream"
  policy      = data.aws_iam_policy_document.videogen-stream-access-description.json
}

resource "aws_iam_policy_attachment" "videogen-lambda-stream-access" {
  name       = "videogen-lambda-stream-access"
  roles      = [aws_iam_role.queue-lambda.name]
  policy_arn = aws_iam_policy.videogen-stream-access.arn
}

data "aws_iam_policy_document" "tts-s3-description" {
  statement {
    actions = ["s3:GetObject"]
    resources = [
      "${aws_s3_bucket.s3_bucket.arn}/assets/*/Summary.txt",
    ]
  }
  statement {
    actions = ["s3:PutObject"]
    resources = [
      "${aws_s3_bucket.s3_bucket.arn}/assets/*/Audio.mp3",
      "${aws_s3_bucket.s3_bucket.arn}/assets/*/Subtitle.ass",
      "${aws_s3_bucket.s3_bucket.arn}/assets/*/TTSResponse.json",
    ]
  }
}

resource "aws_iam_policy" "tts-s3" {
  name        = "tts-s3"
  description = "Allows the subtitle function to read and write to the s3 bucket."
  policy      = data.aws_iam_policy_document.tts-s3-description.json
}

resource "aws_iam_role_policy_attachment" "lambda-tts-s3-access" {
  role       = aws_iam_role.tts-role.name
  policy_arn = aws_iam_policy.tts-s3.arn
}

# CREATE the Queue Lambda Role
resource "aws_iam_role" "queue-lambda" {
  name               = "queue-lambda"
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
  roles      = [aws_iam_role.queue-lambda.name]
  policy_arn = aws_iam_policy.innerVPC-policy.arn
}

# Lambda logging policy attachment
resource "aws_iam_policy_attachment" "submit-cloudwatch" {
  name       = "submit-cloudwatch"
  roles      = [aws_iam_role.submit-job-role.name, aws_iam_role.lambda-authorizer-role.name, aws_iam_role.tts-role.name, aws_iam_role.queue-lambda.name]
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
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

# DEFINE the s3 access policy for the ECS Consumer Task Role
data "aws_iam_policy_document" "ecs-consumer-task-s3" {
  statement {
    effect  = "Allow"
    actions = ["s3:GetObject"]
    resources = [
      "${aws_s3_bucket.s3_bucket.arn}/assets/*/Audio.mp3",
      "${aws_s3_bucket.s3_bucket.arn}/assets/*/Subtitle.ass",
      "${aws_s3_bucket.s3_bucket.arn}/background/*"
    ]
  }
  statement {
    effect  = "Allow"
    actions = ["s3:PutObject"]
    resources = [
      "${aws_s3_bucket.s3_bucket.arn}/assets/*/*.mp4"
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

# DEFINE the dynamodb access policy for the ECS Consumer Task Role
data "aws_iam_policy_document" "ecs-consumer-task-dynamo" {
  statement {
    actions = ["dynamodb:UpdateItem"]
    resources = [
      aws_dynamodb_table.jobs-table.arn,
    ]
  }
}

# CREATE the dynamo access policy for the ECS Consumer Task Role
resource "aws_iam_policy" "ecs-consumer-task-dynamo" {
  name        = "lecture-analyzer-ecs-consumer-task-dynamo"
  description = "Allows the task to access dynamo"
  policy      = data.aws_iam_policy_document.ecs-consumer-task-dynamo.json
}

# ATTACH the dynamo access policy to the ECS Consumer Task Role
resource "aws_iam_role_policy_attachment" "ecs-consumer-task-dynamo-policy" {
  role       = aws_iam_role.ecs-consumer-task-role.name
  policy_arn = aws_iam_policy.ecs-consumer-task-dynamo.arn
}

# DEFINE the sesv2 access policy for the ECS Consumer Task Role
data "aws_iam_policy_document" "ecs-consumer-task-sesv2" {
  statement {
    actions   = ["ses:SendTemplatedEmail"]
    resources = ["*"]
  }
}

# CREATE the sesv2 access policy for the ECS Consumer Task Role
resource "aws_iam_policy" "ecs-consumer-task-sesv2" {
  name        = "lecture-analyzer-ecs-consumer-task-sesv2"
  description = "Allows the task to access SES"
  policy      = data.aws_iam_policy_document.ecs-consumer-task-sesv2.json
}

# ATTACH the sesv2 access policy to the ECS Consumer Task Role
resource "aws_iam_role_policy_attachment" "ecs-consumer-task-sesv2-policy" {
  role       = aws_iam_role.ecs-consumer-task-role.name
  policy_arn = aws_iam_policy.ecs-consumer-task-sesv2.arn
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
