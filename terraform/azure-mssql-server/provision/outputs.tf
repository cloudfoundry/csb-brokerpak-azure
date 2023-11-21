output "sqldbResourceGroup" { value = azurerm_sql_server.azure_sql_db_server.resource_group_name }
output "sqlServerName" { value = azurerm_sql_server.azure_sql_db_server.name }
output "sqlServerFullyQualifiedDomainName" { value = azurerm_sql_server.azure_sql_db_server.fully_qualified_domain_name }
output "hostname" { value = azurerm_sql_server.azure_sql_db_server.fully_qualified_domain_name }
output "port" { value = 1433 }
output "username" { value = local.admin_username }
output "password" {
  value     = local.admin_password
  sensitive = true
}
output "databaseLogin" { value = local.admin_username }
output "databaseLoginPassword" {
  value     = local.admin_password
  sensitive = true
}