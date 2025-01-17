# This file sets up the various IAM roles needed by the application.
# -> ECS Task Role (Applied to EC2 instances)
# -> ECS Task Execution Role (Applied to the ECS service)

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

# CREATE the ECS Task Role
resource "aws_iam_role" "ecs-task-role" {
  name               = "lecture-analyzer-ecs-task-role"
  assume_role_policy = data.aws_iam_policy_document.ecs-task-trust-policy.json
}

# ATTACH the role to the required ECS Task Role policy
resource "aws_iam_role_policy_attachment" "ecs-task-role-policy" {
  role       = aws_iam_role.ecs-task-role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

# CREATE a profile for the ECS Task Role
resource "aws_iam_instance_profile" "ecs-task-profile" {
  name = "lecture-analyzer-ecs-task-profile"
  role = aws_iam_role.ecs-task-role.name
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

# CREATE the ECS Task Execution Role
resource "aws_iam_role" "ecs-task-execution-role" {
  name               = "lecture-analyzer-ecs-task-execution-role"
  assume_role_policy = data.aws_iam_policy_document.ecs-task-execution-trust-policy.json
}

# ATTACH the role to the required ECS Task Execution Role policy
resource "aws_iam_role_policy_attachment" "ecs-task-execution-role-policy" {
  role       = aws_iam_role.ecs-task-execution-role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}
