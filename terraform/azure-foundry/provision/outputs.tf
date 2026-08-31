output "instance_name" {
  value = var.instance_name
}

output "implementation_state" {
  value = "preview_key_backed"
}

output "cf_provenance_json" {
  value = jsonencode(local.cf_provenance)
}

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

output "resource_tags_json" {
  value = jsonencode(local.common_tags)
}

output "account_name" {
  value = azurerm_cognitive_account.openai.name
}

output "foundry_hub_name" {
  value = var.foundry_hub_name
}

output "foundry_project_name" {
  value = var.foundry_project_name
}

output "deployment_name" {
  value = var.deployment_name
}

output "model_name" {
  value = var.model_name
}

output "model_version" {
  value = var.model_version
}

output "model_capacity" {
  value = var.model_capacity
}

output "resource_group" {
  value = azurerm_resource_group.openai.name
}

output "location" {
  value = var.location
}

output "azure_tenant_id" {
  value = var.azure_tenant_id
}

output "azure_subscription_id" {
  value = var.azure_subscription_id
}

output "budget_amount" {
  value = var.budget_amount
}

output "budget_enforcement_mode" {
  value = local.budget_enforcement_mode
}

output "token_audience" {
  value = var.token_audience
}

output "ttl_expires_at" {
  value = local.ttl_expires_at
}