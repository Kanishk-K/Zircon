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
      "${aws_s3_bucket.s3_bucket.arn}/assets/*/Audio.aac",
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

# DEFINE the cognito access policy for the ECS Consumer Task Role
data "aws_iam_policy_document" "ecs-consumer-task-cognito" {
  statement {
    actions   = ["cognito-idp:AdminGetUser"]
    resources = [aws_cognito_user_pool.zircon_user_pool.arn]
  }
}

# CREATE the cognito access policy for the ECS Consumer Task Role
resource "aws_iam_policy" "ecs-consumer-task-cognito" {
  name        = "lecture-analyzer-ecs-consumer-task-cognito"
  description = "Allows the task to access Cognito Emails"
  policy      = data.aws_iam_policy_document.ecs-consumer-task-cognito.json
}

# ATTACH the cognito access policy to the ECS Consumer Task Role
resource "aws_iam_role_policy_attachment" "ecs-consumer-task-cognito-policy" {
  role       = aws_iam_role.ecs-consumer-task-role.name
  policy_arn = aws_iam_policy.ecs-consumer-task-cognito.arn
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
