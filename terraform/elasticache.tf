# This file outlines the deployment of the task queue using ElsatiCache.
# -> ElastiCache subnet group
# -> ElastiCache cluster

# CREATES an ElastiCache subnet group
resource "aws_elasticache_subnet_group" "elasticache-subnet-group" {
  name       = "lecture-analyzer-elasticache-subnet-group"
  subnet_ids = aws_subnet.private-subnets[*].id
}

# CREATES an ElastiCache cluster
resource "aws_elasticache_replication_group" "task-queue" {
  replication_group_id       = "lecture-analyzer-task-queue"
  description                = "Lecture Analyzer Redis Replication Group"
  engine                     = "redis"
  engine_version             = "7.1"
  node_type                  = "cache.t3.micro"
  num_cache_clusters         = 1
  automatic_failover_enabled = false # Set to true for multi-AZ failover
  port                       = 6379
  parameter_group_name       = "default.redis7"
  subnet_group_name          = aws_elasticache_subnet_group.elasticache-subnet-group.name
  security_group_ids         = [aws_security_group.elasticache-sg.id]

  tags = {
    Name        = "lecture-analyzer-task-queue"
    Environment = "prod"
  }
}
