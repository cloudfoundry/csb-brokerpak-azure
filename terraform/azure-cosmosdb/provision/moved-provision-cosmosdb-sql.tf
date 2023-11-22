moved {
  from = module.instance.azurerm_resource_group.rg
  to   = azurerm_resource_group.rg
}
moved {
  from = module.instance.azurerm_cosmosdb_account.cosmosdb-account
  to   = azurerm_cosmosdb_account.cosmosdb-account
}
moved {
  from = module.instance.azurerm_cosmosdb_sql_database.db
  to   = azurerm_cosmosdb_sql_database.db
}
moved {
  from = module.instance.azurerm_private_endpoint.private_endpoint
  to   = azurerm_private_endpoint.private_endpoint
}
