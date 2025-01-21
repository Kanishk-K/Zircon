resource "aws_s3_bucket" "s3_bucket" {
  bucket = "lecture-processor"
  tags = {
    Name        = "lecture-processor"
    Environment = "prod"
  }
  lifecycle {
    prevent_destroy = true
  }
}

resource "aws_s3_bucket_policy" "read_content" {
  bucket = aws_s3_bucket.s3_bucket.bucket
  policy = data.aws_iam_policy_document.read_content.json
}