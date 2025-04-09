# ATTENTION: You'll have to define the styles for the login page in the AWS Cognito Console yourself. 
# Terraform doesn't support this yet.

resource "aws_cognito_user_pool" "zircon_user_pool" {
  name = "zircon-user-pool"
  lambda_config {
    pre_sign_up       = aws_lambda_function.pre_signup_lambda.arn
    post_confirmation = aws_lambda_function.post_signup_lambda.arn
  }
  lifecycle {
    prevent_destroy = true
  }
}

resource "aws_cognito_identity_provider" "zircon_google_oauth" {
  user_pool_id  = aws_cognito_user_pool.zircon_user_pool.id
  provider_name = "Google"
  provider_type = "Google"
  provider_details = {
    client_id        = var.GOOGLE_CLIENT_ID
    client_secret    = var.GOOGLE_CLIENT_SECRET
    authorize_scopes = "email profile"
  }
  attribute_mapping = {
    email = "email"
    name  = "name"
  }
}

resource "aws_cognito_user_pool_domain" "zircon_auth_domain" {
  domain          = "auth.${var.DOMAIN}"
  certificate_arn = aws_acm_certificate.cognito_ssl_cert.arn
  user_pool_id    = aws_cognito_user_pool.zircon_user_pool.id
}

resource "aws_cognito_user_pool_client" "zircon_app_client" {
  name = "zircon-app-client"

  user_pool_id                         = aws_cognito_user_pool.zircon_user_pool.id
  callback_urls                        = ["https://${var.PROD_EXTENSION_ID}.chromiumapp.org/callback", "https://${var.EXTENSION_ID}.chromiumapp.org/callback"]
  supported_identity_providers         = [aws_cognito_identity_provider.zircon_google_oauth.provider_name]
  allowed_oauth_flows                  = ["code"]
  allowed_oauth_scopes                 = ["email", "openid", "profile"]
  allowed_oauth_flows_user_pool_client = true
  id_token_validity                    = 1  # 1 hour
  access_token_validity                = 1  # 1 hour
  refresh_token_validity               = 30 # 30 days
}
