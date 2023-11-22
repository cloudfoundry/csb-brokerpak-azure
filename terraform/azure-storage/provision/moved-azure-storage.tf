moved {
  from = module.instance.random_string.account_name
  to   = random_string.account_name
}
moved {
  from = module.instance.azurerm_resource_group.azure-storage
  to   = azurerm_resource_group.azure-storage
}
moved {
  from = module.instance.azurerm_storage_account.account
  to   = azurerm_storage_account.account
}
moved {
  from = module.instance.azurerm_storage_account_network_rules.account_network_rule
  to   = azurerm_storage_account_network_rules.account_network_rule
}
