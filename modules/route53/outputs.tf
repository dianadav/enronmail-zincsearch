output "zone_id" {
  description = "ID de la zona hospedada en Route 53"
  value       = aws_route53_zone.frontend_zone.zone_id
}

output "domain_name" {
  description = "Nombre del dominio configurado en Route 53"
  value       = aws_route53_zone.frontend_zone.name
}
