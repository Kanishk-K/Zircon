resource "aws_route53_record" "base" {
  zone_id = var.DOMAIN_ZONE_ID
  name    = var.DOMAIN
  type    = "A"
  alias {
    name                   = aws_cloudfront_distribution.web_routing.domain_name
    zone_id                = aws_cloudfront_distribution.web_routing.hosted_zone_id
    evaluate_target_health = false
  }
}

resource "aws_route53_record" "www" {
  zone_id = var.DOMAIN_ZONE_ID
  name    = "www.${var.DOMAIN}"
  type    = "A"
  alias {
    name                   = aws_cloudfront_distribution.web_routing.domain_name
    zone_id                = aws_cloudfront_distribution.web_routing.hosted_zone_id
    evaluate_target_health = false
  }
}

# CREATE the SSL certificate
resource "aws_acm_certificate" "ssl_cert" {
  domain_name = var.DOMAIN
  subject_alternative_names = [
    "www.${var.DOMAIN}"
  ]
  validation_method = "DNS"
  lifecycle {
    create_before_destroy = true
  }
}

# CREATE the Route53 records for the SSL certificate
resource "aws_route53_record" "ssl_cert" {
  for_each = {
    for dvo in aws_acm_certificate.ssl_cert.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      type   = dvo.resource_record_type
      record = dvo.resource_record_value
    }
  }
  name            = each.value.name
  type            = each.value.type
  ttl             = 60
  allow_overwrite = true
  records         = [each.value.record]
  zone_id         = var.DOMAIN_ZONE_ID
}

# VALIDATE the SSL certificate
resource "aws_acm_certificate_validation" "ssl_validation" {
  certificate_arn         = aws_acm_certificate.ssl_cert.arn
  validation_record_fqdns = [for record in aws_route53_record.ssl_cert : record.fqdn]
}

# CREATE the COGNITO SSL certificate
resource "aws_acm_certificate" "cognito_ssl_cert" {
  domain_name       = "auth.${var.DOMAIN}"
  validation_method = "DNS"
  lifecycle {
    create_before_destroy = true
  }
}

# CREATE the Route53 records for the SSL certificate
resource "aws_route53_record" "cognito_ssl_cert" {
  for_each = {
    for dvo in aws_acm_certificate.cognito_ssl_cert.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      type   = dvo.resource_record_type
      record = dvo.resource_record_value
    }
  }
  name            = each.value.name
  type            = each.value.type
  ttl             = 60
  allow_overwrite = true
  records         = [each.value.record]
  zone_id         = var.DOMAIN_ZONE_ID
}

# VALIDATE the SSL certificate
resource "aws_acm_certificate_validation" "cognito_ssl_validation" {
  certificate_arn         = aws_acm_certificate.cognito_ssl_cert.arn
  validation_record_fqdns = [for record in aws_route53_record.cognito_ssl_cert : record.fqdn]
}

# ADD Cognito Record to Route53
resource "aws_route53_record" "auth-cognito" {
  name    = aws_cognito_user_pool_domain.zircon_auth_domain.domain
  type    = "A"
  zone_id = var.DOMAIN_ZONE_ID
  alias {
    name                   = aws_cognito_user_pool_domain.zircon_auth_domain.cloudfront_distribution
    zone_id                = aws_cognito_user_pool_domain.zircon_auth_domain.cloudfront_distribution_zone_id
    evaluate_target_health = false
  }
}

# SES Route Validation
resource "aws_route53_record" "ses_validation" {
  zone_id = var.DOMAIN_ZONE_ID
  type    = "TXT"
  ttl     = "600"
  records = [aws_ses_domain_identity.zircon_domain_identity.verification_token]
  name    = aws_ses_domain_identity.zircon_domain_identity.id
}

# SES DMARC Validation 
resource "aws_route53_record" "dmarc_validation" {
  zone_id = var.DOMAIN_ZONE_ID
  type    = "TXT"
  ttl     = "600"
  records = ["v=DMARC1; p=none; rua=mailto:dmarc-reports@${var.DOMAIN}"]
  name    = "_dmarc.${var.DOMAIN}"
}

resource "aws_route53_record" "dkim_validation" {
  count   = 3
  zone_id = var.DOMAIN_ZONE_ID
  type    = "CNAME"
  ttl     = "300"
  records = ["${aws_ses_domain_dkim.zircon_dkim.dkim_tokens[count.index]}.dkim.amazonses.com"]
  name    = "${aws_ses_domain_dkim.zircon_dkim.dkim_tokens[count.index]}._domainkey.${var.DOMAIN}"
}
