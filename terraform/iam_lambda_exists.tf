resource "aws_iam_role" "exists_lambda_role" {
  name               = "exists-lambda-role"
  assume_role_policy = data.aws_iam_policy_document.lambda-trust-policy.json
}

data "aws_iam_policy_document" "exists_lambda_description" {
  statement {
    actions = ["dynamodb:GetItem"]
    resources = [
      aws_dynamodb_table.jobs-table.arn,
    ]
  }
}

resource "aws_iam_policy" "exists_lambda" {
  name        = "exists-lambda"
  description = "Allows the exists lambda to access the jobs dynamodb table"
  policy      = data.aws_iam_policy_document.exists_lambda_description.json
}

resource "aws_iam_role_policy_attachment" "exists_policy_attachment" {
  role       = aws_iam_role.exists_lambda_role.name
  policy_arn = aws_iam_policy.exists_lambda.arn
}
