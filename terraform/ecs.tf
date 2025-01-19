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
    arn = aws_iam_instance_profile.ecs-task-profile.arn
  }
  user_data = base64encode(templatefile("${path.cwd}/aws/ecs_template.sh", { "cluster" : aws_ecs_cluster.consumer-cluster.name }))
}

# CREATE an ASG for the ECS consumer
resource "aws_autoscaling_group" "ecs-consumer-asg" {
  name                = "lecture-analyzer-ecs-consumer-asg"
  vpc_zone_identifier = aws_subnet.private-subnets[*].id
  min_size            = 1
  max_size            = 1
  desired_capacity    = 1
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
  task_role_arn      = aws_iam_role.ecs-task-role.arn
  execution_role_arn = aws_iam_role.ecs-task-execution-role.arn
  network_mode       = "awsvpc"
  cpu                = 1024
  memory             = 952
  container_definitions = jsonencode([{
    cpu         = 1024
    memory      = 952
    name        = "lecture-analyzer-consumer-container"
    image       = "${aws_ecrpublic_repository.consumer-images.repository_uri}:latest"
    essential   = true
    environment = []
  }])
}

# CREATE a service for the ECS consumer
resource "aws_ecs_service" "consumer-service" {
  name            = "lecture-analyzer-ecs-consumer-service"
  cluster         = aws_ecs_cluster.consumer-cluster.id
  task_definition = aws_ecs_task_definition.ecs-consumer-task-definition.arn
  desired_count   = 1
  capacity_provider_strategy {
    capacity_provider = aws_ecs_capacity_provider.ecs-consumer-capacity-provider.name
    base              = 1
    weight            = 100
  }
  network_configuration {
    subnets         = aws_subnet.private-subnets[*].id
    security_groups = [aws_security_group.ecs-node-sg.id]
  }
  depends_on = [aws_ecs_cluster_capacity_providers.ecs-consumer-capacity-provider]
}

# -> ECS Producer Cluster
# -> ECS Producer Launch Template
# -> ECS Producer ASG
# -> ECS Producer Capacity Provider (and assignment)
# -> ECS Producer Application Load Balancer

# CREATE an ECS cluster for the producer
resource "aws_ecs_cluster" "producer-cluster" {
  name = "lecture-analyzer-producer-cluster"
}

# CREATE a launch template for the ECS producer
resource "aws_launch_template" "ecs-producer-launch-template" {
  name                   = "lecture-analyzer-ecs-producer-launch-template"
  image_id               = var.ecs-ami
  instance_type          = "t2.micro"
  vpc_security_group_ids = [aws_security_group.ecs-node-sg.id]
  iam_instance_profile {
    arn = aws_iam_instance_profile.ecs-task-profile.arn
  }
  user_data = base64encode(templatefile("${path.cwd}/aws/ecs_template.sh", { "cluster" : aws_ecs_cluster.producer-cluster.name }))
}

# CREATE an ASG for the ECS producer
resource "aws_autoscaling_group" "ecs-producer-asg" {
  name                = "lecture-analyzer-ecs-producer-asg"
  vpc_zone_identifier = aws_subnet.public-subnets[*].id
  min_size            = 1
  max_size            = 1
  desired_capacity    = 1
  launch_template {
    id      = aws_launch_template.ecs-producer-launch-template.id
    version = "$Latest"
  }
  tag {
    key                 = "Name"
    value               = "lecture-analyzer-ecs-producer"
    propagate_at_launch = true
  }
  tag {
    key                 = "AmazonECSManaged"
    value               = true
    propagate_at_launch = true
  }
}

# CREATE a (singular) capacity provider for the ECS producer so that ECS can use the containers
resource "aws_ecs_capacity_provider" "ecs-producer-capacity-provider" {
  name = "lecture-analyzer-ecs-producer-capacity-provider"
  auto_scaling_group_provider {
    auto_scaling_group_arn         = aws_autoscaling_group.ecs-producer-asg.arn
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
    Name        = "lecture-analyzer-ecs-producer-capacity-provider"
    Environment = "prod"
  }
}

# ASSIGN the capacity provider to the ECS producer cluster
resource "aws_ecs_cluster_capacity_providers" "ecs-producer-capacity-provider" {
  cluster_name       = aws_ecs_cluster.producer-cluster.name
  capacity_providers = [aws_ecs_capacity_provider.ecs-producer-capacity-provider.name]
  default_capacity_provider_strategy {
    capacity_provider = aws_ecs_capacity_provider.ecs-producer-capacity-provider.name
    base              = 1
    weight            = 100
  }
}

# CREATE a task definition for the ECS producer
resource "aws_ecs_task_definition" "ecs-producer-task-definition" {
  family             = "lecture-analyzer-ecs-producer"
  task_role_arn      = aws_iam_role.ecs-task-role.arn
  execution_role_arn = aws_iam_role.ecs-task-execution-role.arn
  network_mode       = "awsvpc"
  cpu                = 1024
  memory             = 952
  container_definitions = jsonencode([{
    cpu       = 1024
    memory    = 952
    name      = "lecture-analyzer-producer-container"
    image     = "${aws_ecrpublic_repository.producer-images.repository_uri}:latest"
    essential = true
    portMappings = [{
      containerPort = 80
      hostPort      = 80
    }]
    environment = [
      {
        name  = "HOST"
        value = "http://${aws_lb.producer-alb.dns_name}"
      }
    ]
    secrets = [
      {
        name      = "GOOGLE_CLIENT_ID"
        valueFrom = aws_ssm_parameter.GOOGLE_CLIENT_ID.arn
      },
      {
        name      = "GOOGLE_CLIENT_SECRET"
        valueFrom = aws_ssm_parameter.GOOGLE_CLIENT_SECRET.arn
      },
      {
        name      = "JWT_PRIVATE"
        valueFrom = aws_ssm_parameter.JWT_PRIVATE.arn
      }
    ]
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = "ecs-producer"
        "awslogs-region"        = "us-east-1"
        "awslogs-stream-prefix" = "ecs-producer-stream"
      }
    }
  }])
}

# CREATE a service for the ECS producer
resource "aws_ecs_service" "producer-service" {
  name            = "lecture-analyzer-ecs-producer-service"
  cluster         = aws_ecs_cluster.producer-cluster.id
  task_definition = aws_ecs_task_definition.ecs-producer-task-definition.arn
  desired_count   = 1
  capacity_provider_strategy {
    capacity_provider = aws_ecs_capacity_provider.ecs-producer-capacity-provider.name
    base              = 1
    weight            = 100
  }
  network_configuration {
    subnets         = aws_subnet.public-subnets[*].id
    security_groups = [aws_security_group.ecs-node-sg.id, aws_security_group.alb-sg.id]
  }
  load_balancer {
    target_group_arn = aws_lb_target_group.producer-tg.arn
    container_name   = "lecture-analyzer-producer-container"
    container_port   = 80
  }
  depends_on = [aws_ecs_cluster_capacity_providers.ecs-producer-capacity-provider, aws_lb_target_group.producer-tg]
}

# CREATE an application load balancer for the ECS producer
resource "aws_lb" "producer-alb" {
  name               = "lecture-analyzer-producer-alb"
  internal           = false
  load_balancer_type = "application"
  subnets            = aws_subnet.public-subnets[*].id
  security_groups    = [aws_security_group.alb-sg.id]
}

# CREATE a target group for the ECS producer
resource "aws_lb_target_group" "producer-tg" {
  name        = "lecture-analyzer-producer-tg"
  port        = 80
  protocol    = "HTTP"
  vpc_id      = aws_vpc.vpc.id
  target_type = "ip"
  health_check {
    path                = "/health"
    protocol            = "HTTP"
    matcher             = "200"
    interval            = 60
    timeout             = 5
    healthy_threshold   = 3
    unhealthy_threshold = 3
  }
}

# CREATE a listener for the ECS producer (HTTP)
resource "aws_lb_listener" "producer-listener" {
  load_balancer_arn = aws_lb.producer-alb.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type = "redirect"
    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}

# CREATE a listener for the ECS producer (HTTPS)
resource "aws_lb_listener" "producer-https-listener" {
  load_balancer_arn = aws_lb.producer-alb.arn
  port              = 443
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_acm_certificate.ssl_cert.arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.producer-tg.arn
  }
}
