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
