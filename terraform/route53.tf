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
