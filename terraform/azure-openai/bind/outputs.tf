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

output "normalized_binding_json" {
  value = jsonencode({
    version            = "v1"
    provider           = "azure"
    provisioner_family = "azure_openai_key"
    connection_type    = "runtime"
    endpoint = {
      base_url    = var.endpoint
      region      = null
      api_version = var.api_version
    }
    access = {
      mode       = "api_key"
      expires_at = null
    }
    grant = {
      kind                 = "scoped_key"
      least_privilege_unit = "resource"
      allowed_models       = [for deployment in try(jsondecode(var.deployments), []) : try(deployment.model, "") if try(deployment.model, "") != ""]
    }
    credential = {
      format = "api_key"
      inline = {
        api_key = var.api_key
      }
      secret_ref = null
    }
  })
  sensitive = true
}
