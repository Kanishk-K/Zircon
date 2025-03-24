data "aws_iam_policy_document" "videogen-stream-access-description" {
  statement {
    actions = ["dynamodb:GetRecords", "dynamodb:GetShardIterator", "dynamodb:DescribeStream", "dynamodb:ListStreams"]
    resources = [
      aws_dynamodb_table.video_requests_table.stream_arn,
    ]
  }
}

resource "aws_iam_policy" "videogen-stream-access" {
  name        = "videogen-stream-access"
  description = "Allows the TTS lambda to access the videogen dynamodb stream"
  policy      = data.aws_iam_policy_document.videogen-stream-access-description.json
}

resource "aws_iam_policy_attachment" "videogen-lambda-stream-access" {
  name       = "videogen-lambda-stream-access"
  roles      = [aws_iam_role.queue-lambda.name]
  policy_arn = aws_iam_policy.videogen-stream-access.arn
}
