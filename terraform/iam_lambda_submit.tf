resource "aws_iam_role" "submit-job-role" {
  name               = "submit-job-role"
  assume_role_policy = data.aws_iam_policy_document.lambda-trust-policy.json
}

data "aws_iam_policy_document" "submit-dynamodb-description" {
  statement {
    actions = ["dynamodb:PutItem"]
    resources = [
      aws_dynamodb_table.jobs-table.arn,
      aws_dynamodb_table.video_requests_table.arn,
    ]
  }
  statement {
    actions = ["dynamodb:UpdateItem"]
    resources = [
      aws_dynamodb_table.jobs-table.arn,
      aws_dynamodb_table.users-table.arn,
    ]
  }
  statement {
    actions = ["dynamodb:DeleteItem"]
    resources = [
      aws_dynamodb_table.jobs-table.arn,
    ]
  }
}

resource "aws_iam_policy" "submit-dynamodb" {
  name        = "submit-dynamodb"
  description = "Allows the submit job lambda to write to the jobs and user tables"
  policy      = data.aws_iam_policy_document.submit-dynamodb-description.json
}

resource "aws_iam_role_policy_attachment" "lambda-submit-dynamodb" {
  role       = aws_iam_role.submit-job-role.name
  policy_arn = aws_iam_policy.submit-dynamodb.arn
}

data "aws_iam_policy_document" "submit-s3-description" {
  statement {
    actions = ["s3:PutObject"]
    resources = [
      "${aws_s3_bucket.s3_bucket.arn}/assets/*/Summary.txt",
      "${aws_s3_bucket.s3_bucket.arn}/assets/*/Notes.md",
    ]
  }
}

resource "aws_iam_policy" "submit-s3" {
  name        = "submit-s3"
  description = "Allows the submit job lambda to write summaries and notes to the S3 bucket"
  policy      = data.aws_iam_policy_document.submit-s3-description.json
}

resource "aws_iam_role_policy_attachment" "lambda-submit-s3" {
  role       = aws_iam_role.submit-job-role.name
  policy_arn = aws_iam_policy.submit-s3.arn
}
