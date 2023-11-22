moved {
  from = module.instance.random_string.username
  to   = random_string.username
}
moved {
  from = module.instance.random_password.password
  to   = random_password.password
}
moved {
  from = module.instance.azurerm_sql_server.azure_sql_db_server
  to   = azurerm_sql_server.azure_sql_db_server
}
moved {
  from = module.instance.azurerm_sql_virtual_network_rule.allow_subnet_id
  to   = azurerm_sql_virtual_network_rule.allow_subnet_id
}
moved {
  from = module.instance.azurerm_sql_firewall_rule.sql_firewall_rule
  to   = azurerm_sql_firewall_rule.sql_firewall_rule
}
moved {
  from = module.instance.azurerm_resource_group.azure_sql
  to   = azurerm_resource_group.azure_sql
}
