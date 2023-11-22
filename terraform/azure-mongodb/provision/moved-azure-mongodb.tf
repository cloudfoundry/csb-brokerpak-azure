moved {
  from = module.instance.random_string.random
  to   = random_string.random
}
moved {
  from = module.instance.azurerm_resource_group.rg
  to   = azurerm_resource_group.rg
}
moved {
  from = module.instance.azurerm_cosmosdb_account.mongo-account
  to   = azurerm_cosmosdb_account.mongo-account
}
moved {
  from = module.instance.azurerm_cosmosdb_mongo_database.mongo-db
  to   = azurerm_cosmosdb_mongo_database.mongo-db
}
moved {
  from = module.instance.azurerm_cosmosdb_mongo_collection.mongo-collection
  to   = azurerm_cosmosdb_mongo_collection.mongo-collection
}
moved {
  from = module.instance.azurerm_private_endpoint.private_endpoint
  to   = azurerm_private_endpoint.private_endpoint
}
