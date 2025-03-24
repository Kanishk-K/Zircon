# CREATE the Lambda Callback Role to allow writing to the users table
resource "aws_iam_role" "lambda-callback-role" {
  name               = "lambda-callback-role"
  assume_role_policy = data.aws_iam_policy_document.lambda-trust-policy.json
}

data "aws_iam_policy_document" "callback-dynamodb-description" {
  statement {
    actions = ["dynamodb:PutItem"]
    resources = [
      aws_dynamodb_table.users-table.arn,
    ]
  }
}

resource "aws_iam_policy" "callback-dynamodb" {
  name        = "callback-dynamodb"
  description = "Allows the callback lambda to write to the users table"
  policy      = data.aws_iam_policy_document.callback-dynamodb-description.json
}

resource "aws_iam_role_policy_attachment" "lambda-callback-dynamodb" {
  role       = aws_iam_role.lambda-callback-role.name
  policy_arn = aws_iam_policy.callback-dynamodb.arn
}
