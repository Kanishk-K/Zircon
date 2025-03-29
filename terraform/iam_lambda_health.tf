resource "aws_iam_role" "health_lambda" {
  name               = "health_lambda"
  assume_role_policy = data.aws_iam_policy_document.lambda-trust-policy.json
}
