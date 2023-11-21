output "sqldbName" { value = azurerm_mssql_database.azure_sql_db.name }
output "sqlServerName" { value = azurerm_sql_failover_group.failover_group.name }
output "sqlServerFullyQualifiedDomainName" { value = local.serverFQDN }
output "hostname" { value = local.serverFQDN }
output "port" { value = 1433 }
output "name" { value = azurerm_mssql_database.azure_sql_db.name }
output "username" { value = random_string.username.result }
output "password" {
  value     = random_password.password.result
  sensitive = true
}
output "status" { value = format("created failover group %s (id: %s), primary db %s (id: %s) on server %s (id: %s), secondary db %s (id: %s/databases/%s) on server %s (id: %s) URL: https://portal.azure.com/#@%s/resource%s/failoverGroup",
  azurerm_sql_failover_group.failover_group.name, azurerm_sql_failover_group.failover_group.id,
  azurerm_mssql_database.azure_sql_db.name, azurerm_mssql_database.azure_sql_db.id,
  azurerm_sql_server.primary_azure_sql_db_server.name, azurerm_sql_server.primary_azure_sql_db_server.id,
  azurerm_mssql_database.azure_sql_db.name, azurerm_sql_server.secondary_sql_db_server.id, azurerm_mssql_database.azure_sql_db.name,
  azurerm_sql_server.secondary_sql_db_server.name, azurerm_sql_server.secondary_sql_db_server.id,
  var.azure_tenant_id,
  azurerm_sql_server.primary_azure_sql_db_server.id) }