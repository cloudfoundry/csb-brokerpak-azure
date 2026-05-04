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

output "budget_amount" {
  value = var.budget_amount
}

output "budget_enforcement_mode" {
  value = var.budget_enforcement_mode
}
