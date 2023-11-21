output "cosmosdb_host_endpoint" { value = azurerm_cosmosdb_account.cosmosdb-account.endpoint }
output "cosmosdb_master_key" {
  value     = azurerm_cosmosdb_account.cosmosdb-account.primary_key
  sensitive = true
}
output "cosmosdb_readonly_master_key" {
  value     = azurerm_cosmosdb_account.cosmosdb-account.primary_readonly_key
  sensitive = true
}
output "cosmosdb_database_id" { value = azurerm_cosmosdb_sql_database.db.name }
output "status" { value = format("created account %s (id: %s) URL: https://portal.azure.com/#@%s/resource%s",
  azurerm_cosmosdb_account.cosmosdb-account.name,
  azurerm_cosmosdb_account.cosmosdb-account.id,
  var.azure_tenant_id,
azurerm_cosmosdb_account.cosmosdb-account.id) }