resource "aws_route53_zone" "frontend_zone" {
  name = var.domain_name
}

resource "aws_route53_record" "frontend_record" {
  zone_id = aws_route53_zone.frontend_zone.zone_id
  name    = var.domain_name
  type    = "A"
  ttl     = 300
  records = ["127.0.0.1"]  # Simula una IP donde está el frontend
}
