output "sqldbResourceGroup" { value = azurerm_sql_server.azure_sql_db_server.resource_group_name }
output "sqldbName" { value = azurerm_sql_database.azure_sql_db.name }
output "sqlServerName" { value = azurerm_sql_server.azure_sql_db_server.name }
output "sqlServerFullyQualifiedDomainName" { value = azurerm_sql_server.azure_sql_db_server.fully_qualified_domain_name }
output "hostname" { value = azurerm_sql_server.azure_sql_db_server.fully_qualified_domain_name }
output "port" { value = 1433 }
output "name" { value = azurerm_sql_database.azure_sql_db.name }
output "username" { value = random_string.username.result }
output "password" {
  value     = random_password.password.result
  sensitive = true
}
output "status" { value = format("created db %s (id: %s) on server %s (id: %s) URL: https://portal.azure.com/#@%s/resource%s",
  azurerm_sql_database.azure_sql_db.name,
  azurerm_sql_database.azure_sql_db.id,
  azurerm_sql_server.azure_sql_db_server.name,
  azurerm_sql_server.azure_sql_db_server.id,
  var.azure_tenant_id,
  azurerm_sql_database.azure_sql_db.id) }