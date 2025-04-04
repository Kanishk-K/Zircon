resource "aws_iam_role" "tts-role" {
  name               = "tts-role"
  assume_role_policy = data.aws_iam_policy_document.lambda-trust-policy.json
}

data "aws_iam_policy_document" "jobs-stream-access-description" {
  statement {
    actions = ["dynamodb:GetRecords", "dynamodb:GetShardIterator", "dynamodb:DescribeStream", "dynamodb:ListStreams"]
    resources = [
      aws_dynamodb_table.jobs-table.stream_arn,
    ]
  }
}

resource "aws_iam_policy" "jobs-stream-access" {
  name        = "jobs-stream-access"
  description = "Allows the TTS lambda to access the jobs dynamodb stream"
  policy      = data.aws_iam_policy_document.jobs-stream-access-description.json
}

resource "aws_iam_policy_attachment" "jobs-lambda-stream-access" {
  name       = "jobs-lambda-stream-access"
  roles      = [aws_iam_role.tts-role.name]
  policy_arn = aws_iam_policy.jobs-stream-access.arn
}

data "aws_iam_policy_document" "tts-s3-description" {
  statement {
    actions = ["s3:GetObject"]
    resources = [
      "${aws_s3_bucket.s3_bucket.arn}/assets/*/Summary.txt",
    ]
  }
  statement {
    actions = ["s3:PutObject"]
    resources = [
      "${aws_s3_bucket.s3_bucket.arn}/assets/*/Audio.aac",
      "${aws_s3_bucket.s3_bucket.arn}/assets/*/Subtitle.ass",
      "${aws_s3_bucket.s3_bucket.arn}/assets/*/TTSResponse.json",
    ]
  }
}

resource "aws_iam_policy" "tts-s3" {
  name        = "tts-s3"
  description = "Allows the subtitle function to read and write to the s3 bucket."
  policy      = data.aws_iam_policy_document.tts-s3-description.json
}

resource "aws_iam_role_policy_attachment" "lambda-tts-s3-access" {
  role       = aws_iam_role.tts-role.name
  policy_arn = aws_iam_policy.tts-s3.arn
}
