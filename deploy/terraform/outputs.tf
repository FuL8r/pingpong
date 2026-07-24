output "service_url" {
  description = "Default App Runner HTTPS URL"
  value       = "https://${aws_apprunner_service.pingpong.service_url}"
}

output "service_arn" {
  description = "App Runner service ARN"
  value       = aws_apprunner_service.pingpong.arn
}

output "ecr_repository_url" {
  description = "ECR repository URL for docker push"
  value       = aws_ecr_repository.pingpong.repository_url
}

output "custom_domain_status" {
  description = "Custom domain association status"
  value       = var.custom_domain != "" ? aws_apprunner_custom_domain_association.pingpong[0].status : ""
}
