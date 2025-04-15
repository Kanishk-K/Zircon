# This file sets up the ECS cluster and task definitions for the application.
# -> ECS Consumer Cluster
# -> ECS Consumer Launch Template
# -> ECS Consumer ASG
# -> ECS Consumer Capacity Provider (and assignment)
# -> ECS Consumer Task Definition
# -> ECS Consumer Service

variable "ecs-ami" {
  description = "This is the AMI that will be used for the ECS instances"
  type        = string
  default     = "ami-0d2df9a9165d36365"
}

# CREATE an ECS cluster for the consumer
resource "aws_ecs_cluster" "consumer-cluster" {
  name = "lecture-analyzer-consumer-cluster"
}

# CREATE a launch template for the ECS consumer
resource "aws_launch_template" "ecs-consumer-launch-template" {
  name                   = "lecture-analyzer-ecs-consumer-launch-template"
  image_id               = var.ecs-ami
  instance_type          = "t2.micro"
  vpc_security_group_ids = [aws_security_group.ecs-node-sg.id]
  iam_instance_profile {
    arn = aws_iam_instance_profile.ecs-consumer-task-profile.arn
  }
  user_data = base64encode(templatefile("${path.cwd}/aws/ecs_template.sh", { "cluster" : aws_ecs_cluster.consumer-cluster.name }))
}

# CREATE an ASG for the ECS consumer
resource "aws_autoscaling_group" "ecs-consumer-asg" {
  name                = "lecture-analyzer-ecs-consumer-asg"
  vpc_zone_identifier = aws_subnet.private-subnets[*].id
  min_size            = 3
  max_size            = 3
  desired_capacity    = 3
  launch_template {
    id      = aws_launch_template.ecs-consumer-launch-template.id
    version = "$Latest"
  }
  tag {
    key                 = "Name"
    value               = "lecture-analyzer-ecs-consumer"
    propagate_at_launch = true
  }
  tag {
    key                 = "AmazonECSManaged"
    value               = true
    propagate_at_launch = true
  }
}

# CREATE a (singular) capacity provider for the ECS consumer so that ECS can use the containers
resource "aws_ecs_capacity_provider" "ecs-consumer-capacity-provider" {
  name = "lecture-analyzer-ecs-consumer-capacity-provider"
  auto_scaling_group_provider {
    auto_scaling_group_arn         = aws_autoscaling_group.ecs-consumer-asg.arn
    managed_termination_protection = "DISABLED"

    managed_scaling {
      maximum_scaling_step_size = 1
      minimum_scaling_step_size = 1
      status                    = "ENABLED"
      target_capacity           = 100
    }
    managed_draining = "ENABLED"
  }
  tags = {
    Name        = "lecture-analyzer-ecs-consumer-capacity-provider"
    Environment = "prod"
  }
}

# ASSIGN the capacity provider to the ECS consumer cluster
resource "aws_ecs_cluster_capacity_providers" "ecs-consumer-capacity-provider" {
  cluster_name       = aws_ecs_cluster.consumer-cluster.name
  capacity_providers = [aws_ecs_capacity_provider.ecs-consumer-capacity-provider.name]
  default_capacity_provider_strategy {
    capacity_provider = aws_ecs_capacity_provider.ecs-consumer-capacity-provider.name
    base              = 1
    weight            = 100
  }
}

# CREATE a task definition for the ECS consumer
resource "aws_ecs_task_definition" "ecs-consumer-task-definition" {
  family             = "lecture-analyzer-ecs-consumer"
  task_role_arn      = aws_iam_role.ecs-consumer-task-role.arn
  execution_role_arn = aws_iam_role.ecs-consumer-task-execution-role.arn
  network_mode       = "host"
  cpu                = 1024
  memory             = 952
  container_definitions = jsonencode([{
    cpu       = 1024
    memory    = 952
    name      = "lecture-analyzer-consumer-container"
    image     = "${aws_ecrpublic_repository.consumer-images.repository_uri}:latest"
    essential = true
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = "ecs-consumer"
        "awslogs-region"        = "us-east-1"
        "awslogs-stream-prefix" = "ecs-consumer-stream"
      }
    }
    environment = [
      {
        name  = "REDIS_URL"
        value = "${aws_elasticache_replication_group.task-queue.primary_endpoint_address}:6379"
      },
      {
        name  = "DOMAIN"
        value = var.DOMAIN
      },
      {
        name  = "COGNITO_POOL"
        value = aws_cognito_user_pool.zircon_user_pool.id
      }
    ]
  }])
}

# CREATE a service for the ECS consumer
resource "aws_ecs_service" "consumer-service" {
  name            = "lecture-analyzer-ecs-consumer-service"
  cluster         = aws_ecs_cluster.consumer-cluster.id
  task_definition = aws_ecs_task_definition.ecs-consumer-task-definition.arn
  desired_count   = 3
  capacity_provider_strategy {
    capacity_provider = aws_ecs_capacity_provider.ecs-consumer-capacity-provider.name
    base              = 1
    weight            = 100
  }
  depends_on = [
    aws_ecs_cluster_capacity_providers.ecs-consumer-capacity-provider,
    aws_s3_object.static_logo,
    aws_s3_object.static_minecraft,
    aws_s3_object.static_subway_surfers
  ]
}
