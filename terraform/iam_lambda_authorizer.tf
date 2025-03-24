# CREATE the Lambda Authorizer Role for protecting the API Gateway
resource "aws_iam_role" "lambda-authorizer-role" {
  name               = "lambda-authorizer-role"
  assume_role_policy = data.aws_iam_policy_document.lambda-trust-policy.json
}
