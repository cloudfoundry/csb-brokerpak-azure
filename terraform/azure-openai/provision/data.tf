locals {
  ttl_hours               = coalesce(var.ttl_hours, 8)
  ttl_expires_at          = timeadd(timestamp(), "${local.ttl_hours}h")
  api_version             = "2024-02-01"
  budget_start_date       = format("%s-01T00:00:00Z", formatdate("YYYY-MM", timestamp()))
  deployments_list        = jsondecode(var.deployments_json)
  cf_context              = try(jsondecode(var.cf_context_json), {})
  cf_originating_identity = try(jsondecode(var.cf_originating_identity_json), {})
  cf_user_id              = try(local.cf_originating_identity.user_id, local.cf_originating_identity.value.user_id, "")
  deployments_by_name = {
    for deployment in local.deployments_list : deployment.name => deployment
  }

  # Truncate instance_name to satisfy Azure 24-char Cognitive Services limit.
  account_name = substr(replace(var.instance_name, "-", ""), 0, 24)

  budget_contact_email      = can(regex("^[^[:space:]@]+@[^[:space:]@]+[.][^[:space:]@]+$", trimspace(var.budget_contact_email))) ? trimspace(var.budget_contact_email) : ""
  budget_webhook_url        = can(regex("^https?://", trimspace(var.budget_webhook_url))) ? trimspace(var.budget_webhook_url) : ""
  budget_contact_emails     = local.budget_contact_email == "" ? [] : [local.budget_contact_email]
  budget_notifications_on   = length(local.budget_contact_emails) > 0 || local.budget_webhook_url != ""
  budget_enforcement_mode   = !local.budget_notifications_on ? "disabled" : local.budget_webhook_url == "" ? "alert-only" : "webhook-automation"
  budget_action_group_name  = substr("${local.account_name}-budget-ag", 0, 64)
  budget_action_group_short = substr(local.account_name, 0, 12)
  budget_name               = substr("${local.account_name}-budget", 0, 63)

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

  deployments = jsonencode(local.deployments_list)
}
