# This file sets up the container registry for the application.
# -> Consumer ECR
# -> Producer ECR

# CREATE an ECR repository for the consumer
resource "aws_ecrpublic_repository" "consumer-images" {
  repository_name = "lecture-analyzer-consumer-images"
}
