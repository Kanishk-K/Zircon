resource "aws_route53_record" "base" {
  zone_id = var.DOMAIN_ZONE_ID
  name    = var.DOMAIN
  type    = "A"
  alias {
    name                   = aws_lb.producer-alb.dns_name
    zone_id                = aws_lb.producer-alb.zone_id
    evaluate_target_health = false
  }
}

resource "aws_route53_record" "www" {
  zone_id = var.DOMAIN_ZONE_ID
  name    = "www.${var.DOMAIN}"
  type    = "A"
  alias {
    name                   = aws_lb.producer-alb.dns_name
    zone_id                = aws_lb.producer-alb.zone_id
    evaluate_target_health = false
  }
}
