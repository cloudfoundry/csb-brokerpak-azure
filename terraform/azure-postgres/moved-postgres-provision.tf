moved {
  from = module.instance.azurerm_resource_group.azure-postgres
  to   = azurerm_resource_group.azure-postgres
}
moved {
  from = module.instance.random_string.username
  to   = random_string.username
}
moved {
  from = module.instance.random_password.password
  to   = random_password.password
}
moved {
  from = module.instance.azurerm_postgresql_server.instance
  to   = azurerm_postgresql_server.instance
}
moved {
  from = module.instance.azurerm_postgresql_database.instance-db
  to   = azurerm_postgresql_database.instance-db
}
moved {
  from = module.instance.azurerm_postgresql_virtual_network_rule.allow_subnet_id
  to   = azurerm_postgresql_virtual_network_rule.allow_subnet_id
}
moved {
  from = module.instance.azurerm_postgresql_firewall_rule.allow_azure
  to   = azurerm_postgresql_firewall_rule.allow_azure
}

