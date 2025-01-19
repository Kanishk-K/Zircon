# This file sets up the various IAM roles needed by the application.
# -> General Role Setup
# -> ECS Producer Task Role (Applied to EC2 producer instances)
# -> ECS Producer Task Execution Role (Applied to the ECS producer service)
# -> ECS Consumer Task Role (Applied to EC2 consumer instances)
# -> ECS Consumer Task Execution Role (Applied to the ECS consumer service)

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

#
# Producer Setup
#

# CREATE the ECS Producer Task Role
resource "aws_iam_role" "ecs-producer-task-role" {
  name               = "lecture-analyzer-ecs-producer-task-role"
  assume_role_policy = data.aws_iam_policy_document.ecs-task-trust-policy.json
}

# ATTACH the required policy to the ECS Producer Task Role
resource "aws_iam_role_policy_attachment" "ecs-producer-task-role-policy" {
  role       = aws_iam_role.ecs-producer-task-role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

# CREATE a profile for the ECS Producer Task Role
resource "aws_iam_instance_profile" "ecs-producer-task-profile" {
  name = "lecture-analyzer-ecs-producer-task-profile"
  role = aws_iam_role.ecs-producer-task-role.name
}

# DEFINE the dynamodb access policy for the ECS Producer Task Role
data "aws_iam_policy_document" "ecs-producer-task-dynamodb" {
  statement {
    actions = ["dynamodb:PutItem", "dynamodb:GetItem"]
    resources = [
      aws_dynamodb_table.jobs-table.arn,
      aws_dynamodb_table.users-table.arn
    ]
  }
}

# CREATE the dynamodb access policy for the ECS Producer Task Role
resource "aws_iam_policy" "ecs-producer-task-dynamodb" {
  name        = "lecture-analyzer-ecs-producer-task-dynamodb"
  description = "Allows the task to access DynamoDB"
  policy      = data.aws_iam_policy_document.ecs-producer-task-dynamodb.json
}

# ATTACH the dynamodb access policy to the ECS Producer Task Role
resource "aws_iam_role_policy_attachment" "ecs-producer-task-policy" {
  role       = aws_iam_role.ecs-producer-task-role.name
  policy_arn = aws_iam_policy.ecs-producer-task-dynamodb.arn
}

# CREATE the ECS Producer Task Execution Role
resource "aws_iam_role" "ecs-producer-task-execution-role" {
  name               = "lecture-analyzer-ecs-producer-task-execution-role"
  assume_role_policy = data.aws_iam_policy_document.ecs-task-execution-trust-policy.json
}

# ATTACH the role to the required ECS Producer Task Execution Role policy
resource "aws_iam_role_policy_attachment" "ecs-producer-task-execution-role-policy" {
  role       = aws_iam_role.ecs-producer-task-execution-role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# DEFINE the secrets access policy for the ECS Producer Task Execution Role
data "aws_iam_policy_document" "ecs-producer-task-execution-secrets" {
  statement {
    actions = ["ssm:GetParameter", "ssm:GetParameters"]
    resources = [
      aws_ssm_parameter.GOOGLE_CLIENT_ID.arn,
      aws_ssm_parameter.GOOGLE_CLIENT_SECRET.arn,
      aws_ssm_parameter.JWT_PRIVATE.arn
    ]
  }
}

# CREATE the ECS Producer Task Execution Role resource access policy
resource "aws_iam_policy" "ecs-producer-task-execution-secrets" {
  name        = "lecture-analyzer-ecs-producer-task-execution-secrets"
  description = "Allows the ECS controller to access resources"
  policy      = data.aws_iam_policy_document.ecs-producer-task-execution-secrets.json
}

# ATTACH the secrets access policy to the ECS Producer Task Execution Role
resource "aws_iam_role_policy_attachment" "ecs-producer-task-execution-policy" {
  role       = aws_iam_role.ecs-producer-task-execution-role.name
  policy_arn = aws_iam_policy.ecs-producer-task-execution-secrets.arn
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
