resource "azurerm_resource_group" "openai" {
  name     = var.resource_group
  location = var.location
  tags     = local.common_tags
}

resource "azurerm_monitor_action_group" "budget" {
  count               = local.budget_webhook_url == "" ? 0 : 1
  name                = local.budget_action_group_name
  resource_group_name = azurerm_resource_group.openai.name
  short_name          = local.budget_action_group_short

  webhook_receiver {
    name                    = "budget-webhook"
    service_uri             = local.budget_webhook_url
    use_common_alert_schema = true
  }

  tags = local.common_tags
}

resource "azurerm_consumption_budget_resource_group" "openai" {
  count             = local.budget_notifications_on ? 1 : 0
  name              = local.budget_name
  resource_group_id = azurerm_resource_group.openai.id
  amount            = var.budget_amount
  time_grain        = "Monthly"

  time_period {
    start_date = local.budget_start_date
  }

  dynamic "notification" {
    for_each = local.budget_notifications_on ? {
      forecast = {
        threshold      = 80
        threshold_type = "Forecasted"
      }
      actual = {
        threshold      = 100
        threshold_type = "Actual"
      }
    } : {}

    content {
      enabled        = true
      operator       = "GreaterThan"
      threshold      = notification.value.threshold
      threshold_type = notification.value.threshold_type
      contact_emails = local.budget_contact_emails
      contact_groups = azurerm_monitor_action_group.budget[*].id
    }
  }
}

# Azure Cognitive Services account with OpenAI kind.
resource "azurerm_cognitive_account" "openai" {
  name                  = local.account_name
  location              = azurerm_resource_group.openai.location
  resource_group_name   = azurerm_resource_group.openai.name
  kind                  = "OpenAI"
  sku_name              = var.sku_name
  custom_subdomain_name = local.account_name
  tags                  = local.common_tags
}

resource "azurerm_cognitive_deployment" "models" {
  for_each             = local.deployments_by_name
  name                 = each.value.name
  cognitive_account_id = azurerm_cognitive_account.openai.id

  model {
    format  = "OpenAI"
    name    = each.value.model
    version = each.value.version
  }

  sku {
    name     = "Standard"
    capacity = tonumber(each.value.capacity)
  }
}
