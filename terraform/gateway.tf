# This file sets up the routing for the API Gateway.
# -> API Gateway
# -> Stage Setup
# -> -> Auth Route
# -> -> Auth Integration
# -> -> Auth Permission
# -> -> API Permissions

resource "aws_apigatewayv2_api" "zircon-api" {
  name          = "zircon-api"
  protocol_type = "HTTP"
  cors_configuration {
    allow_methods = ["GET", "POST"]
    allow_origins = ["*"]
    allow_headers = [
      "Authorization"
    ]
  }
}

resource "aws_apigatewayv2_stage" "zircon-stage" {
  api_id      = aws_apigatewayv2_api.zircon-api.id
  name        = "zircon-stage"
  auto_deploy = true
}

# Login Route
resource "aws_apigatewayv2_route" "login-route" {
  api_id             = aws_apigatewayv2_api.zircon-api.id
  route_key          = "GET /login"
  authorization_type = "NONE"
  target             = "integrations/${aws_apigatewayv2_integration.login-integration.id}"
}

resource "aws_apigatewayv2_integration" "login-integration" {
  api_id             = aws_apigatewayv2_api.zircon-api.id
  integration_type   = "AWS_PROXY"
  connection_type    = "INTERNET"
  integration_method = "POST"
  integration_uri    = aws_lambda_function.login-lambda.invoke_arn
}

resource "aws_lambda_permission" "login-integration-perm" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.login-lambda.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.zircon-api.execution_arn}/*"
}

# Callback Route
resource "aws_apigatewayv2_route" "callback-route" {
  api_id             = aws_apigatewayv2_api.zircon-api.id
  route_key          = "GET /callback"
  authorization_type = "NONE"
  target             = "integrations/${aws_apigatewayv2_integration.callback-integration.id}"
}

resource "aws_apigatewayv2_integration" "callback-integration" {
  api_id             = aws_apigatewayv2_api.zircon-api.id
  integration_type   = "AWS_PROXY"
  connection_type    = "INTERNET"
  integration_method = "POST"
  integration_uri    = aws_lambda_function.callback-lambda.invoke_arn
}

resource "aws_lambda_permission" "callback-integration-perm" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.callback-lambda.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.zircon-api.execution_arn}/*"
}

# Authorizer
resource "aws_apigatewayv2_authorizer" "lambda-authorizer" {
  api_id                            = aws_apigatewayv2_api.zircon-api.id
  authorizer_type                   = "REQUEST"
  identity_sources                  = ["$request.header.Authorization"]
  name                              = "lambda-authorizer"
  authorizer_uri                    = aws_lambda_function.auth-lambda.invoke_arn
  enable_simple_responses           = true
  authorizer_payload_format_version = "2.0"
}

resource "aws_lambda_permission" "auth-invocation-perm" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.auth-lambda.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.zircon-api.execution_arn}/*"
}

# Submit Job Route
resource "aws_apigatewayv2_route" "submit-job-route" {
  api_id             = aws_apigatewayv2_api.zircon-api.id
  route_key          = "POST /submitJob"
  authorization_type = "CUSTOM"
  authorizer_id      = aws_apigatewayv2_authorizer.lambda-authorizer.id
  target             = "integrations/${aws_apigatewayv2_integration.submit-job-integration.id}"
}

resource "aws_apigatewayv2_integration" "submit-job-integration" {
  api_id             = aws_apigatewayv2_api.zircon-api.id
  integration_type   = "AWS_PROXY"
  connection_type    = "INTERNET"
  integration_method = "POST"
  integration_uri    = aws_lambda_function.job-lambda.invoke_arn
}

resource "aws_lambda_permission" "submit-job-integration-perm" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.job-lambda.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.zircon-api.execution_arn}/*"
}
