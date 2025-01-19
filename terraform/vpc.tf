# This file sets up the networking configurations for the application.
# -> VPC Setup
# -> Subnet Setup
# -> Internet Gateway Setup
# -> VPC Endpoint (Gateways) Setup
# -> EIP Setup
# -> NAT Gateway Setup
# -> Route Table Setup
# -> Security Group Setup


# CREATES a VPC for the application with a CIDR block of 10.0.0.0/16 (65,536 IP addresses)
resource "aws_vpc" "vpc" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name        = "lecture-analyzer"
    Environment = "prod"
  }
}

# DEFINES public subnets for the application each with around 256 IP addresses
variable "public_subnet_cidr_blocks" {
  type        = list(string)
  description = "These are the CIDR blocks that will be used to generate public subnets"
  default     = ["10.0.1.0/24", "10.0.2.0/24"]
}

# DEFINES private subnets for the application each with around 256 IP addresses
variable "private_subnet_cidr_blocks" {
  type        = list(string)
  description = "These are the CIDR blocks that will be used to generate private subnets"
  default     = ["10.0.3.0/24"]
}

# DEFINES availability zones for the application will use. Ensure the length of this list is equal to the number of subnets.
variable "azs" {
  type        = list(string)
  description = "These are the availability zones that will be used to generate subnets"
  default     = ["us-east-1a", "us-east-1b"]
}

# CREATES public subnets for the application
resource "aws_subnet" "public-subnets" {
  count                   = length(var.public_subnet_cidr_blocks)
  vpc_id                  = aws_vpc.vpc.id
  cidr_block              = element(var.public_subnet_cidr_blocks, count.index)
  availability_zone       = element(var.azs, count.index)
  map_public_ip_on_launch = true

  tags = {
    "Name" = "lecture-analyzer-public-subnet-${count.index}"
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
    "Name" = "lecture-analyzer-private-subnet-${count.index}"
  }
}

# CREATES an internet gateway for the VPC to allow internet access
resource "aws_internet_gateway" "igw" {
  vpc_id = aws_vpc.vpc.id

  tags = {
    Name        = "lecture-analyzer-igw"
    Environment = "prod"
  }
}

# CREATES a EIP for the NAT gateway
resource "aws_eip" "eip" {
  domain = "vpc"
  tags = {
    Name        = "lecture-analyzer-eip"
    Environment = "prod"
  }
}

# CREATES a NAT gateway for the private subnets to allow internet access
resource "aws_nat_gateway" "nat-gateway" {
  allocation_id = aws_eip.eip.id
  subnet_id     = element(aws_subnet.public-subnets[*].id, 0)
  depends_on    = [aws_internet_gateway.igw]

  tags = {
    Name        = "lecture-analyzer-nat-gateway"
    Environment = "prod"
  }
}

# CREATES a VPC endpoint to allow private access to S3
resource "aws_vpc_endpoint" "s3-endpoint" {
  vpc_id          = aws_vpc.vpc.id
  service_name    = "com.amazonaws.us-east-1.s3"
  route_table_ids = [aws_route_table.rt-public.id, aws_route_table.rt-private.id]
  tags = {
    Name        = "lecture-analyzer-s3-endpoint"
    Environment = "prod"
  }
}

# CREATES a VPC endpoint to allow private access to DynamoDB
resource "aws_vpc_endpoint" "dynamodb-endpoint" {
  vpc_id          = aws_vpc.vpc.id
  service_name    = "com.amazonaws.us-east-1.dynamodb"
  route_table_ids = [aws_route_table.rt-public.id, aws_route_table.rt-private.id]
  tags = {
    Name        = "lecture-analyzer-dynamodb-endpoint"
    Environment = "prod"
  }
}

# CREATES a public route table for the public subnets
resource "aws_route_table" "rt-public" {
  vpc_id = aws_vpc.vpc.id
  tags = {
    Name        = "lecture-analyzer-rt-public"
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
    Name        = "lecture-analyzer-rt-private"
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
  name   = "lecture-analyzer-ecs-node-sg"
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
}

# CREATES a security group for the ALB to allow traffic from the internet
resource "aws_security_group" "alb-sg" {
  name   = "lecture-analyzer-alb-sg"
  vpc_id = aws_vpc.vpc.id

  ingress {
    # HTTP
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  ingress {
    # HTTPS
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  egress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# CREATES a security group for the producer service to communicate with the ALB and Oauth service
resource "aws_security_group" "producer-sg" {
  name   = "lecture-analyzer-producer-sg"
  vpc_id = aws_vpc.vpc.id

  ingress {
    # OAuth, this is fine as we have no listeners for this port.
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  ingress {
    # ALB, only allow port 80 traffic from the ALB as we do have a listener for this port.
    from_port       = 80
    to_port         = 80
    protocol        = "tcp"
    security_groups = [aws_security_group.alb-sg.id]
  }
}
