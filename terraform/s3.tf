resource "aws_s3_bucket" "s3_bucket" {
  bucket = "lecture-processor"
  tags = {
    Name        = "lecture-processor"
    Environment = "prod"
  }
}
