# This file sets up the networking configurations for the application.
# -> VPC Setup
# -> Subnet Setup
# -> Internet Gateway Setup
# -> VPC Endpoint (Gateways) Setup
# -> EIP Setup
# -> NAT Gateway Setup
# -> Route Table Setup
# -> Security Group Setup
# -> -> ECS Task Security Group
# -> -> ALB Security Group
# -> -> Producer Security Group
# -> -> ElastiCache Security Group


# CREATES a VPC for the application with a CIDR block of 10.0.0.0/16 (65,536 IP addresses)
resource "aws_vpc" "vpc" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name        = "zircon-vpc"
    Environment = "prod"
  }
}

# DEFINES public subnets for the application each with around 256 IP addresses
variable "public_subnet_cidr_blocks" {
  type        = list(string)
  description = "These are the CIDR blocks that will be used to generate public subnets"
  default     = ["10.0.1.0/24"]
}

# DEFINES private subnets for the application each with around 256 IP addresses
variable "private_subnet_cidr_blocks" {
  type        = list(string)
  description = "These are the CIDR blocks that will be used to generate private subnets"
  default     = ["10.0.2.0/24"]
}

# DEFINES availability zones for the application will use. Ensure the length of this list is equal to the number of subnets.
variable "azs" {
  type        = list(string)
  description = "These are the availability zones that will be used to generate subnets"
  default     = ["us-east-1a"]
}

# CREATES public subnets for the application
resource "aws_subnet" "public-subnets" {
  count                   = length(var.public_subnet_cidr_blocks)
  vpc_id                  = aws_vpc.vpc.id
  cidr_block              = element(var.public_subnet_cidr_blocks, count.index)
  availability_zone       = element(var.azs, count.index)
  map_public_ip_on_launch = true

  tags = {
    "Name" = "zircon-public-subnet-${count.index}"
  }
}

# CREATES private subnets for the application
resource "aws_subnet" "private-subnets" {
  count                   = length(var.private_subnet_cidr_blocks)
  vpc_id                  = aws_vpc.vpc.id
  cidr_block              = element(var.private_subnet_cidr_blocks, count.index)
  availability_zone       = element(var.azs, count.index)
  map_public_ip_on_launch = false

  tags = {
    "Name" = "zircon-private-subnet-${count.index}"
  }
}

# CREATES an internet gateway for the VPC to allow internet access
resource "aws_internet_gateway" "igw" {
  vpc_id = aws_vpc.vpc.id

  tags = {
    Name        = "zircon-igw"
    Environment = "prod"
  }
}

# CREATES a EIP for the NAT gateway
resource "aws_eip" "eip" {
  domain = "vpc"
  tags = {
    Name        = "zircon-eip"
    Environment = "prod"
  }
}

# CREATES a NAT gateway for the private subnets to allow internet access
resource "aws_nat_gateway" "nat-gateway" {
  allocation_id = aws_eip.eip.id
  subnet_id     = element(aws_subnet.public-subnets[*].id, 0)
  depends_on    = [aws_internet_gateway.igw]

  tags = {
    Name        = "zircon-nat-gateway"
    Environment = "prod"
  }
}

# CREATES a VPC endpoint to allow private access to S3
resource "aws_vpc_endpoint" "s3-endpoint" {
  vpc_id          = aws_vpc.vpc.id
  service_name    = "com.amazonaws.us-east-1.s3"
  route_table_ids = [aws_route_table.rt-public.id, aws_route_table.rt-private.id]
  tags = {
    Name        = "zircon-s3-endpoint"
    Environment = "prod"
  }
}

# CREATES a VPC endpoint to allow private access to DynamoDB
resource "aws_vpc_endpoint" "dynamodb-endpoint" {
  vpc_id          = aws_vpc.vpc.id
  service_name    = "com.amazonaws.us-east-1.dynamodb"
  route_table_ids = [aws_route_table.rt-public.id, aws_route_table.rt-private.id]
  tags = {
    Name        = "zircon-dynamodb-endpoint"
    Environment = "prod"
  }
}

# CREATES a public route table for the public subnets
resource "aws_route_table" "rt-public" {
  vpc_id = aws_vpc.vpc.id
  tags = {
    Name        = "zircon-rt-public"
    Environment = "prod"
  }
}

# CREATES a route to the internet gateway for the public route table
resource "aws_route" "rt-public-route" {
  route_table_id         = aws_route_table.rt-public.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.igw.id
}

# ASSIGN public subnets to the public route table
resource "aws_route_table_association" "rt-public-assoc" {
  count          = length(aws_subnet.public-subnets)
  subnet_id      = aws_subnet.public-subnets[count.index].id
  route_table_id = aws_route_table.rt-public.id
}

# CREATES a private route table for the private subnets
resource "aws_route_table" "rt-private" {
  vpc_id = aws_vpc.vpc.id
  tags = {
    Name        = "zircon-rt-private"
    Environment = "prod"
  }
}

# CREATES a route to the NAT gateway for the private route table
resource "aws_route" "rt-private-route" {
  route_table_id         = aws_route_table.rt-private.id
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = aws_nat_gateway.nat-gateway.id
}

# ASSIGN private subnets to the private route table
resource "aws_route_table_association" "rt-private-assoc" {
  count          = length(aws_subnet.private-subnets)
  subnet_id      = aws_subnet.private-subnets[count.index].id
  route_table_id = aws_route_table.rt-private.id
}

# CREATES a security group for the nodes to pull from the ECR repository and communicate with the ECS cluster
# https://repost.aws/questions/QUk9Za0ev-Rzeas-FoTHbymQ/security-group-outbound-rules-with-elastic-container-service
resource "aws_security_group" "ecs-node-sg" {
  name   = "zircon-ecs-node-sg"
  vpc_id = aws_vpc.vpc.id

  # Allow traffic to the ECS cluster
  egress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  egress {
    from_port   = 53
    to_port     = 53
    protocol    = "udp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  egress {
    # Docker ports
    from_port   = 2375
    to_port     = 2376
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  egress {
    from_port   = 51678
    to_port     = 51680
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  egress {
    from_port       = 6379
    to_port         = 6379
    protocol        = "tcp"
    security_groups = [aws_security_group.elasticache-sg.id]
  }
}

resource "aws_security_group" "elasticache-sg" {
  name        = "zircon-elasticache-sg"
  description = "Allow traffic to the ElastiCache cluster"
  vpc_id      = aws_vpc.vpc.id

  ingress {
    from_port   = 6379
    to_port     = 6379
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "lambda-elasticache-sg" {
  name        = "zircon-lambda-elasticache-sg"
  description = "Allow traffic to the ElastiCache cluster from a Lambda function"
  vpc_id      = aws_vpc.vpc.id
  egress {
    from_port       = 6379
    to_port         = 6379
    protocol        = "tcp"
    security_groups = [aws_security_group.elasticache-sg.id]
  }
}
