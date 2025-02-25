output "uri" {
  value     = replace(azurerm_cosmosdb_account.mongo-account.primary_mongodb_connection_string, "/?", "/${azurerm_cosmosdb_mongo_database.mongo-db.name}?")
  sensitive = true
}
output "status" {
  value = format(
    "created db %s (id: %s) URL:  https://portal.azure.com/#@%s/resource%s",
    azurerm_cosmosdb_mongo_database.mongo-db.name,
    azurerm_cosmosdb_mongo_database.mongo-db.id,
    var.azure_tenant_id,
    azurerm_cosmosdb_mongo_database.mongo-db.id,
  )
}