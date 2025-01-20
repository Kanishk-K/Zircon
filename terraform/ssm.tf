resource "aws_ssm_parameter" "GOOGLE_CLIENT_ID" {
  name  = "/lecture-analyzer/google_client_id"
  type  = "String"
  value = var.GOOGLE_CLIENT_ID
}

resource "aws_ssm_parameter" "GOOGLE_CLIENT_SECRET" {
  name  = "/lecture-analyzer/google_client_secret"
  type  = "String"
  value = var.GOOGLE_CLIENT_SECRET
}

resource "aws_ssm_parameter" "JWT_PRIVATE" {
  name  = "/lecture-analyzer/jwt_private"
  type  = "String"
  value = var.JWT_PRIVATE
}

resource "aws_ssm_parameter" "OPENAI_API_KEY" {
  name  = "/lecture-analyzer/openai_api_key"
  type  = "String"
  value = var.OPENAI_API_KEY
}