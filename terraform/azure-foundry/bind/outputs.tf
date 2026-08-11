output "instance_name" {
  value = var.instance_name
}

output "implementation_state" {
  value = var.implementation_state
}

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

output "binding_provenance_json" {
  value = jsonencode({
    cf_organization_guid = try(jsondecode(var.cf_context_json).organization_guid, "")
    cf_organization_name = try(jsondecode(var.cf_context_json).organization_name, "")
    cf_space_guid        = try(jsondecode(var.cf_context_json).space_guid, "")
    cf_space_name        = try(jsondecode(var.cf_context_json).space_name, "")
    cf_user_id           = try(jsondecode(var.cf_originating_identity_json).user_id, jsondecode(var.cf_originating_identity_json).value.user_id, "")
    cf_app_guid          = var.cf_app_guid
  })
}

output "foundry_hub_name" {
  value = var.foundry_hub_name
}

output "foundry_project_name" {
  value = var.foundry_project_name
}

output "resource_tags_json" {
  value = var.resource_tags_json
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
  value = var.resource_group
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

output "budget_enforcement_mode" {
  value = var.budget_enforcement_mode
}

output "token_audience" {
  value = var.token_audience
}

output "ttl_expires_at" {
  value = var.ttl_expires_at
}

output "normalized_binding_json" {
  value = jsonencode({
    version            = "v1"
    provider           = "azure"
    provisioner_family = "azure_foundry_identity"
    connection_type    = "runtime"
    endpoint = {
      base_url    = var.endpoint
      region      = var.location
      api_version = var.api_version
    }
    access = {
      mode       = "api_key"
      expires_at = var.ttl_expires_at
    }
    grant = {
      kind                 = "scoped_key"
      least_privilege_unit = "model"
      allowed_models       = [var.model_name]
    }
    credential = {
      format = "api_key"
      inline = {
        api_key         = var.api_key
        deployment_name = var.deployment_name
      }
      secret_ref = null
    }
  })
  sensitive = true
}