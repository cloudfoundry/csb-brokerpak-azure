locals {
  ttl_expires_at          = timeadd(timestamp(), "${var.ttl_hours}h")
  cf_context              = try(jsondecode(var.cf_context_json), {})
  cf_originating_identity = try(jsondecode(var.cf_originating_identity_json), {})
  cf_user_id              = try(local.cf_originating_identity.user_id, local.cf_originating_identity.value.user_id, "")

  cf_provenance = {
    cf_organization_guid = try(local.cf_context.organization_guid, "")
    cf_organization_name = try(local.cf_context.organization_name, "")
    cf_space_guid        = try(local.cf_context.space_guid, "")
    cf_space_name        = try(local.cf_context.space_name, "")
    cf_user_id           = local.cf_user_id
  }

  common_tags = merge(var.labels, local.cf_provenance, {
    TTLExpiry   = local.ttl_expires_at
    ManagedBy   = "cloud-service-broker"
    Environment = "sandbox"
  })

  cluster_name        = substr(var.instance_name, 0, 63)
  vnet_name           = substr("${var.instance_name}-vnet", 0, 64)
  subnet_name         = substr("${var.instance_name}-nodes", 0, 80)
  node_resource_group = substr("${var.resource_group}-nodes", 0, 80)
  api_server          = "https://${var.private_cluster_enabled ? azurerm_kubernetes_cluster.aks.private_fqdn : azurerm_kubernetes_cluster.aks.fqdn}"
}
