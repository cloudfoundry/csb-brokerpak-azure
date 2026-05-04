output "endpoint" {
  value = var.endpoint
}

output "api_key" {
  value     = var.api_key
  sensitive = true
}

output "api_version" {
  value = var.api_version
}

output "deployments" {
  value = var.deployments
}

output "ttl_expires_at" {
  value = var.ttl_expires_at
}
