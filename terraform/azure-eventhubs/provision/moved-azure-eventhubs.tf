moved {
  from = module.instance.azurerm_eventhub.eventhub
  to   = azurerm_eventhub.eventhub
}
moved {
  from = module.instance.azurerm_resource_group.rg
  to   = azurerm_resource_group.rg
}
moved {
  from = module.instance.azurerm_eventhub_namespace.rg-namespace
  to   = azurerm_eventhub_namespace.rg-namespace
}
