resource "aws_iam_role" "queue-lambda" {
  name               = "queue-lambda"
  assume_role_policy = data.aws_iam_policy_document.lambda-trust-policy.json
}
