resource "aws_iam_role" "post_signup_lambda_role" {
  name               = "post-signup-lambda-role"
  assume_role_policy = data.aws_iam_policy_document.lambda-trust-policy.json
}

data "aws_iam_policy_document" "post_signup_lambda_description" {
  statement {
    actions = ["dynamodb:PutItem"]
    resources = [
      aws_dynamodb_table.users-table.arn,
    ]
  }
}

resource "aws_iam_policy" "post_signup_lambda" {
  name        = "post-signup-lambda"
  description = "Allows the post signup lambda to write to the users table"
  policy      = data.aws_iam_policy_document.post_signup_lambda_description.json
}

resource "aws_iam_role_policy_attachment" "post_signup_policy_attachment" {
  role       = aws_iam_role.post_signup_lambda_role.name
  policy_arn = aws_iam_policy.post_signup_lambda.arn
}
