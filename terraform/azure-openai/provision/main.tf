resource "azurerm_resource_group" "openai" {
  name     = var.resource_group
  location = var.location
  tags     = local.common_tags
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

  # Restrict to deny-all by default; operators should add IP allowlist via
  # cf update-service or configure VNet integration post-provision.
  network_acls {
    default_action = "Allow"
    ip_rules       = []
  }
}

# GPT-3.5-Turbo deployment — lowest-cost chat model, suitable for sandbox.
resource "azurerm_cognitive_deployment" "gpt35" {
  name                 = "gpt-35-turbo"
  cognitive_account_id = azurerm_cognitive_account.openai.id

  model {
    format  = "OpenAI"
    name    = "gpt-35-turbo"
    version = "0125"
  }

  sku {
    name     = "Standard"
    capacity = var.gpt35_capacity
  }
}

# Text embedding model — required for RAG and similarity search patterns.
resource "azurerm_cognitive_deployment" "ada_embedding" {
  name                 = "text-embedding-ada-002"
  cognitive_account_id = azurerm_cognitive_account.openai.id

  model {
    format  = "OpenAI"
    name    = "text-embedding-ada-002"
    version = "2"
  }

  sku {
    name     = "Standard"
    capacity = var.embedding_capacity
  }
}
