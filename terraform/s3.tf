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

resource "aws_s3_object" "static_logo" {
  bucket      = aws_s3_bucket.s3_bucket.bucket
  key         = "background/logo.png"
  source      = "${path.module}/../backend/static/logo.png"
  source_hash = filemd5("${path.module}/../backend/static/logo.png")
}

resource "aws_s3_object" "static_minecraft" {
  bucket      = aws_s3_bucket.s3_bucket.bucket
  key         = "background/minecraft.mp4"
  source      = "${path.module}/../backend/static/minecraft.mp4"
  source_hash = filemd5("${path.module}/../backend/static/minecraft.mp4")
}

resource "aws_s3_object" "static_subway_surfers" {
  bucket      = aws_s3_bucket.s3_bucket.bucket
  key         = "background/subway_surfers.mp4"
  source      = "${path.module}/../backend/static/subway_surfers.mp4"
  source_hash = filemd5("${path.module}/../backend/static/subway_surfers.mp4")
}

resource "aws_s3_object" "static_baking" {
  bucket      = aws_s3_bucket.s3_bucket.bucket
  key         = "background/baking.mp4"
  source      = "${path.module}/../backend/static/baking.mp4"
  source_hash = filemd5("${path.module}/../backend/static/baking.mp4")
}

resource "aws_s3_object" "static_makeup" {
  bucket      = aws_s3_bucket.s3_bucket.bucket
  key         = "background/makeup.mp4"
  source      = "${path.module}/../backend/static/makeup.mp4"
  source_hash = filemd5("${path.module}/../backend/static/makeup.mp4")
}

resource "aws_s3_object" "static_sand" {
  bucket      = aws_s3_bucket.s3_bucket.bucket
  key         = "background/sand.mp4"
  source      = "${path.module}/../backend/static/sand.mp4"
  source_hash = filemd5("${path.module}/../backend/static/sand.mp4")
}

resource "aws_s3_object" "static_soap" {
  bucket      = aws_s3_bucket.s3_bucket.bucket
  key         = "background/soap.mp4"
  source      = "${path.module}/../backend/static/soap.mp4"
  source_hash = filemd5("${path.module}/../backend/static/soap.mp4")
}
