resource "aws_iam_role" "ttl-role" {
  name               = "ttl-role"
  assume_role_policy = data.aws_iam_policy_document.lambda-trust-policy.json
}

data "aws_iam_policy_document" "video-request-stream-access-description" {
  statement {
    actions = ["dynamodb:GetRecords", "dynamodb:GetShardIterator", "dynamodb:DescribeStream", "dynamodb:ListStreams"]
    resources = [
      aws_dynamodb_table.video_requests_table.stream_arn,
    ]
  }
}

resource "aws_iam_policy" "video-request-stream-access" {
  name        = "video-request-stream-access"
  description = "Allows the TTL lambda to access the video request dynamodb stream"
  policy      = data.aws_iam_policy_document.video-request-stream-access-description.json
}

resource "aws_iam_policy_attachment" "video-request-lambda-stream-access" {
  name       = "video-request-lambda-stream-access"
  roles      = [aws_iam_role.ttl-role.name]
  policy_arn = aws_iam_policy.video-request-stream-access.arn
}

data "aws_iam_policy_document" "ttl-s3-description" {
  statement {
    actions = ["s3:DeleteObject"]
    resources = [
      "${aws_s3_bucket.s3_bucket.arn}/assets/*/*.mp4",
    ]
  }
}

resource "aws_iam_policy" "ttl-s3" {
  name        = "ttl-s3"
  description = "Allows the TTL lambda to delete objects from S3"
  policy      = data.aws_iam_policy_document.ttl-s3-description.json
}

resource "aws_iam_policy_attachment" "ttl-s3-attachment" {
  name       = "ttl-s3-attachment"
  roles      = [aws_iam_role.ttl-role.name]
  policy_arn = aws_iam_policy.ttl-s3.arn
}

data "aws_iam_policy_document" "ttl-dynamo-jobs-description" {
  statement {
    actions = ["dynamodb:UpdateItem"]
    resources = [
      aws_dynamodb_table.jobs-table.arn,
    ]
  }
}

resource "aws_iam_policy" "ttl-dynamo-jobs" {
  name        = "ttl-dynamo-jobs"
  description = "Allows the TTL lambda to remove elments from the videosAvailable attribute in the Jobs table."
  policy      = data.aws_iam_policy_document.ttl-dynamo-jobs-description.json
}

resource "aws_iam_policy_attachment" "ttl-dynamo-jobs-attachment" {
  name       = "ttl-dynamo-jobs-attachment"
  roles      = [aws_iam_role.ttl-role.name]
  policy_arn = aws_iam_policy.ttl-dynamo-jobs.arn
}
