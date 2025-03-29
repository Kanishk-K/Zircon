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

# Authorizer
resource "aws_apigatewayv2_authorizer" "cognito_authorizer" {
  name             = "cognito-authorizer"
  api_id           = aws_apigatewayv2_api.zircon-api.id
  authorizer_type  = "JWT"
  identity_sources = ["$request.header.Authorization"]
  jwt_configuration {
    audience = [aws_cognito_user_pool_client.zircon_app_client.id]
    issuer   = "https://${aws_cognito_user_pool.zircon_user_pool.endpoint}"
  }

}

# Submit Job Route
resource "aws_apigatewayv2_route" "submit-job-route" {
  api_id             = aws_apigatewayv2_api.zircon-api.id
  route_key          = "POST /submitJob"
  authorization_type = "JWT"
  authorizer_id      = aws_apigatewayv2_authorizer.cognito_authorizer.id
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

# Exists Route
resource "aws_apigatewayv2_route" "exists-route" {
  api_id             = aws_apigatewayv2_api.zircon-api.id
  route_key          = "GET /exists"
  authorization_type = "JWT"
  authorizer_id      = aws_apigatewayv2_authorizer.cognito_authorizer.id
  target             = "integrations/${aws_apigatewayv2_integration.exists-integration.id}"
}

resource "aws_apigatewayv2_integration" "exists-integration" {
  api_id             = aws_apigatewayv2_api.zircon-api.id
  integration_type   = "AWS_PROXY"
  connection_type    = "INTERNET"
  integration_method = "POST"
  integration_uri    = aws_lambda_function.exists_lambda.invoke_arn
}

resource "aws_lambda_permission" "exists-integration-perm" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.exists_lambda.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.zircon-api.execution_arn}/*"
}

# Health Route
resource "aws_apigatewayv2_route" "health-route" {
  api_id    = aws_apigatewayv2_api.zircon-api.id
  route_key = "GET /health"
  target    = "integrations/${aws_apigatewayv2_integration.health-integration.id}"
}

resource "aws_apigatewayv2_integration" "health-integration" {
  api_id             = aws_apigatewayv2_api.zircon-api.id
  integration_type   = "AWS_PROXY"
  connection_type    = "INTERNET"
  integration_method = "POST"
  integration_uri    = aws_lambda_function.health_lambda.invoke_arn
}

resource "aws_lambda_permission" "health-integration-perm" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.health_lambda.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.zircon-api.execution_arn}/*"
}
