resource "aws_apprunner_custom_domain_association" "pingpong" {
  count       = var.custom_domain != "" ? 1 : 0
  domain_name = var.custom_domain
  service_arn = aws_apprunner_service.pingpong.arn
}

# Certificate-validation CNAMEs (created only when DNS is in Route53)
resource "aws_route53_record" "validation" {
  for_each = (var.custom_domain != "" && var.hosted_zone_id != "") ? {
    for r in aws_apprunner_custom_domain_association.pingpong[0].certificate_validation_records :
    r.name => r
  } : {}

  zone_id         = var.hosted_zone_id
  name            = each.value.name
  type            = each.value.type
  records         = [each.value.value]
  ttl             = 300
  allow_overwrite = true
}

# CNAME from the custom domain to the App Runner DNS target
# NOTE App Runner's target is a CNAME - use a subdomain (ping.example.com)
resource "aws_route53_record" "target" {
  count   = (var.custom_domain != "" && var.hosted_zone_id != "") ? 1 : 0
  zone_id = var.hosted_zone_id
  name    = var.custom_domain
  type    = "CNAME"
  records = [aws_apprunner_custom_domain_association.pingpong[0].dns_target]
  ttl     = 300
}
