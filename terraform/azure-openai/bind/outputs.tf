output "endpoint" {
  value = var.endpoint
}

output "instance_name" {
  value = var.instance_name
}

output "resource_tags_json" {
  value = var.resource_tags_json
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

output "budget_enforcement_mode" {
  value = var.budget_enforcement_mode
}
