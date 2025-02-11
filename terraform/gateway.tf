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
    allow_origins = ["https://${var.DOMAIN}", "https://www.${var.DOMAIN}"]
  }
}

resource "aws_apigatewayv2_stage" "zircon-stage" {
  api_id      = aws_apigatewayv2_api.zircon-api.id
  name        = "zircon-stage"
  auto_deploy = true
}

resource "aws_apigatewayv2_route" "auth-route" {
  api_id             = aws_apigatewayv2_api.zircon-api.id
  route_key          = "GET /auth"
  authorization_type = "NONE"
  target             = "integrations/${aws_apigatewayv2_integration.auth-integration.id}"
}

resource "aws_apigatewayv2_integration" "auth-integration" {
  api_id             = aws_apigatewayv2_api.zircon-api.id
  integration_type   = "AWS_PROXY"
  connection_type    = "INTERNET"
  integration_method = "POST"
  integration_uri    = aws_lambda_function.auth-lambda.invoke_arn
}

resource "aws_lambda_permission" "auth-integration-perm" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.auth-lambda.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.zircon-api.execution_arn}/*"
}



