moved {
  from = module.instance.azurerm_resource_group.azure-sql-fog
  to   = azurerm_resource_group.azure-sql-fog
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
  from = module.instance.azurerm_sql_server.primary_azure_sql_db_server
  to   = azurerm_sql_server.primary_azure_sql_db_server
}
moved {
  from = module.instance.azurerm_sql_server.secondary_sql_db_server
  to   = azurerm_sql_server.secondary_sql_db_server
}
moved {
  from = module.instance.azurerm_mssql_database.secondary_azure_sql_db
  to   = azurerm_mssql_database.secondary_azure_sql_db
}
moved {
  from = module.instance.azurerm_mssql_database.azure_sql_db
  to   = azurerm_mssql_database.azure_sql_db
}
moved {
  from = module.instance.azurerm_sql_failover_group.failover_group
  to   = azurerm_sql_failover_group.failover_group
}
moved {
  from = module.instance.azurerm_sql_virtual_network_rule.allow_subnet_id2
  to   = azurerm_sql_virtual_network_rule.allow_subnet_id2
}
moved {
  from = module.instance.azurerm_sql_virtual_network_rule.allow_subnet_id1
  to   = azurerm_sql_virtual_network_rule.allow_subnet_id1
}
moved {
  from = module.instance.azurerm_sql_firewall_rule.server1
  to   = azurerm_sql_firewall_rule.server1
}
moved {
  from = module.instance.azurerm_sql_firewall_rule.server2
  to   = azurerm_sql_firewall_rule.server2
}
