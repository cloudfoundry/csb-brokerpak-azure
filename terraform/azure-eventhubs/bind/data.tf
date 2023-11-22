data "azurerm_eventhub_namespace" "ns" {
  name                = var.namespace_name
  resource_group_name = var.eventhub_rg_name
}
