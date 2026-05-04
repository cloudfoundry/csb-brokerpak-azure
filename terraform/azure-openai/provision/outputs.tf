output "endpoint" {
  value = azurerm_cognitive_account.openai.endpoint
}

output "api_key" {
  value     = azurerm_cognitive_account.openai.primary_access_key
  sensitive = true
}

output "api_version" {
  value = local.api_version
}

output "deployments" {
  value = local.deployments
}

output "account_name" {
  value = azurerm_cognitive_account.openai.name
}

output "resource_group" {
  value = azurerm_resource_group.openai.name
}

output "ttl_expires_at" {
  value = local.ttl_expires_at
}
