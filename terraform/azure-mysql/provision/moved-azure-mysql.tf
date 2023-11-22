moved {
  from = module.instance.azurerm_resource_group.azure-msyql
  to   = azurerm_resource_group.azure-msyql
}
moved {
  from = module.instance.random_string.username
  to   = random_string.username
}
moved {
  from = module.instance.random_string.servername
  to   = random_string.servername
}
moved {
  from = module.instance.random_password.password
  to   = random_password.password
}
moved {
  from = module.instance.random_string.random
  to   = random_string.random
}
moved {
  from = module.instance.azurerm_mysql_server.instance
  to   = azurerm_mysql_server.instance
}
moved {
  from = module.instance.azurerm_mysql_database.instance-db
  to   = azurerm_mysql_database.instance-db
}
moved {
  from = module.instance.azurerm_mysql_virtual_network_rule.allow_subnet_id
  to   = azurerm_mysql_virtual_network_rule.allow_subnet_id
}
moved {
  from = module.instance.azurerm_mysql_virtual_network_rule.allow_subnet_ids
  to   = azurerm_mysql_virtual_network_rule.allow_subnet_ids
}
moved {
  from = module.instance.azurerm_mysql_firewall_rule.allow_azure
  to   = azurerm_mysql_firewall_rule.allow_azure
}
moved {
  from = module.instance.azurerm_mysql_firewall_rule.allow_firewall
  to   = azurerm_mysql_firewall_rule.allow_firewall
}
moved {
  from = module.instance.azurerm_private_endpoint.private_endpoint
  to   = azurerm_private_endpoint.private_endpoint
}

