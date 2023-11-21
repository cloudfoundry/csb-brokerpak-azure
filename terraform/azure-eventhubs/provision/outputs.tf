output "eventhub_rg_name" { value = local.resource_group }
output "namespace_name" { value = azurerm_eventhub_namespace.rg-namespace.name }
output "eventhub_name" { value = azurerm_eventhub.eventhub.name }
output "status" { value = format("created event hub %s (id: %s)  URL: https://portal.azure.com/#@%s/resource%s",
  azurerm_eventhub.eventhub.name,
  azurerm_eventhub.eventhub.id,
  var.azure_tenant_id,
azurerm_eventhub.eventhub.id) }