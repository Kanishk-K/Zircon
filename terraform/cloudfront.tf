# This file sets up the CloudFront distribution for the application.
# -> Cloudfront Origin Access Control
# -> CloudFront Distribution

locals {
  s3_origin_id    = "zircon-s3-origin"
  apigw_origin_id = "zircon-apigw-origin"
}

# CREATE a CloudFront Origin Access Control
resource "aws_cloudfront_origin_access_control" "s3_access_control" {
  name                              = "lecture-analyzer-s3-access-control"
  description                       = "Restrict access to the S3 bucket"
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}

# CREATE a CloudFront distribution
resource "aws_cloudfront_distribution" "web_routing" {
  origin {
    domain_name = replace(aws_apigatewayv2_stage.zircon-stage.invoke_url, "/^https?://([^/]*).*/", "$1")
    origin_id   = local.apigw_origin_id
    origin_path = "/${aws_apigatewayv2_stage.zircon-stage.name}"
    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }
  origin {
    domain_name              = aws_s3_bucket.s3_bucket.bucket_regional_domain_name
    origin_id                = local.s3_origin_id
    origin_access_control_id = aws_cloudfront_origin_access_control.s3_access_control.id
  }
  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD", "OPTIONS", "PUT", "POST", "PATCH", "DELETE"]
    cached_methods         = ["GET", "HEAD", "OPTIONS"]
    target_origin_id       = local.apigw_origin_id
    viewer_protocol_policy = "redirect-to-https"
    compress               = true
    forwarded_values {
      query_string = true # Needed for oauth
      headers = [
        "Authorization"
      ]
      cookies {
        forward = "none"
      }
    }
  }
  ordered_cache_behavior {
    path_pattern           = "/assets/*"
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = local.s3_origin_id
    viewer_protocol_policy = "redirect-to-https"
    compress               = true
    forwarded_values {
      query_string = false
      cookies {
        forward = "none"
      }
    }
    default_ttl = 86400
    max_ttl     = 31536000
    min_ttl     = 0
  }
  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA"]
    }
  }
  viewer_certificate {
    acm_certificate_arn      = aws_acm_certificate.ssl_cert.arn
    ssl_support_method       = "sni-only"
    minimum_protocol_version = "TLSv1.2_2021"
  }
  enabled     = true
  aliases     = [var.DOMAIN, "www.${var.DOMAIN}"]
  price_class = "PriceClass_100"
  tags = {
    Name        = "lecture-analyzer-s3-distribution"
    Environment = "prod"
  }
  depends_on = [aws_acm_certificate.ssl_cert]
}
