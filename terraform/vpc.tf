# This file sets up the networking configurations for the application.
# -> VPC Setup
# -> Subnet Setup
# -> Route Table Setup
# -> Internet Gateway Setup
# -> VPC Endpoint (Gateways) Setup
# -> NAT Gateway Setup
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
  default     = ["10.0.1.0/24"]
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
  default     = ["us-east-1a"]
}

# CREATES public subnets for the application
resource "aws_subnet" "public-subnets" {
  count             = length(var.public_subnet_cidr_blocks)
  vpc_id            = aws_vpc.vpc.id
  cidr_block        = element(var.public_subnet_cidr_blocks, count.index)
  availability_zone = element(var.azs, count.index)

  tags = {
    "Name" = "lecture-analyzer-public-subnet-${count.index}"
  }
}

# CREATES private subnets for the application
resource "aws_subnet" "private-subnets" {
  count             = length(var.private_subnet_cidr_blocks)
  vpc_id            = aws_vpc.vpc.id
  cidr_block        = element(var.private_subnet_cidr_blocks, count.index)
  availability_zone = element(var.azs, count.index)

  tags = {
    "Name" = "lecture-analyzer-private-subnet-${count.index}"
  }
}

# CREATES a public route table for the public subnets
resource "aws_route_table" "rt-public" {
  vpc_id = aws_vpc.vpc.id

  route {
    # Send all external traffic to the internet gateway
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.igw.id
  }
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
}

# ASSIGN private subnets to the private route table
resource "aws_route_table_association" "rt-private-assoc" {
  count          = length(aws_subnet.private-subnets)
  subnet_id      = aws_subnet.private-subnets[count.index].id
  route_table_id = aws_route_table.rt-private.id
}

# CREATES an internet gateway for the VPC to allow internet access
resource "aws_internet_gateway" "igw" {
  vpc_id = aws_vpc.vpc.id

  tags = {
    Name        = "lecture-analyzer-igw"
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

# CREATES a NAT gateway to allow private subnets to access the internet
