locals {
  ttl_expires_at = timeadd(timestamp(), "${var.ttl_hours}h")
  api_version    = "2024-02-01"
  budget_start_date = format("%s-01T00:00:00Z", formatdate("YYYY-MM", timestamp()))
  deployments_list = jsondecode(var.deployments_json)
  deployments_by_name = {
    for deployment in local.deployments_list : deployment.name => deployment
  }

  # Truncate instance_name to satisfy Azure 24-char Cognitive Services limit.
  account_name = substr(replace(var.instance_name, "-", ""), 0, 24)

  budget_contact_emails      = trimspace(var.budget_contact_email) == "" ? [] : [trimspace(var.budget_contact_email)]
  budget_notifications_on    = length(local.budget_contact_emails) > 0 || trimspace(var.budget_webhook_url) != ""
  budget_enforcement_mode    = trimspace(var.budget_webhook_url) == "" ? "alert-only" : "webhook-automation"
  budget_action_group_name   = substr("${local.account_name}-budget-ag", 0, 64)
  budget_action_group_short  = substr(local.account_name, 0, 12)
  budget_name                = substr("${local.account_name}-budget", 0, 63)

  common_tags = merge(var.labels, {
    TTLExpiry   = local.ttl_expires_at
    ManagedBy   = "cloud-service-broker"
    Environment = "sandbox"
  })

  deployments = jsonencode(local.deployments_list)
}
