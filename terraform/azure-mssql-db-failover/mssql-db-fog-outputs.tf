output "sqldbName" { value = var.existing ? var.db_name : one(azurerm_mssql_database.primary_db[*].name) }
output "sqlServerName" { value = var.existing ? var.instance_name : one(azurerm_sql_failover_group.failover_group[*].name) }
output "sqlServerFullyQualifiedDomainName" { value = format("%s.database.windows.net", var.existing ? var.instance_name : one(azurerm_sql_failover_group.failover_group[*].name)) }
output "hostname" { value = format("%s.database.windows.net", var.existing ? var.instance_name : one(azurerm_sql_failover_group.failover_group[*].name)) }
output "port" { value = 1433 }
output "name" { value = var.existing ? var.db_name : one(azurerm_mssql_database.primary_db[*].name) }
output "username" { value = var.server_credential_pairs[var.server_pair].admin_username }
output "password" { value = var.server_credential_pairs[var.server_pair].admin_password }
output "server_pair" { value = var.server_pair }
output "status" {
  value = var.existing ? format("connected to existing failover group - primary server %s (id: %s) secondary server %s (%s) URL: https://portal.azure.com/#@%s/resource%s/failoverGroup",
    data.azurerm_sql_server.primary_sql_db_server.name, data.azurerm_sql_server.primary_sql_db_server.id,
    data.azurerm_sql_server.secondary_sql_db_server.name, data.azurerm_sql_server.secondary_sql_db_server.id,
    var.azure_tenant_id,
    data.azurerm_sql_server.primary_sql_db_server.id) : format("created failover group %s (id: %s), primary db %s (id: %s) on server %s (id: %s), secondary db %s (id: %s/databases/%s) on server %s (id: %s) URL: https://portal.azure.com/#@%s/resource%s/failoverGroup",
    one(azurerm_sql_failover_group.failover_group[*].name), one(azurerm_sql_failover_group.failover_group[*].id),
    one(azurerm_sql_failover_group.failover_group[*].name), one(azurerm_mssql_database.primary_db[*].id),
    data.azurerm_sql_server.primary_sql_db_server.name, data.azurerm_sql_server.primary_sql_db_server.id,
    one(azurerm_mssql_database.primary_db[*].name), data.azurerm_sql_server.secondary_sql_db_server.id, one(azurerm_mssql_database.primary_db[*].name),
    data.azurerm_sql_server.secondary_sql_db_server.name, data.azurerm_sql_server.secondary_sql_db_server.id,
    var.azure_tenant_id,
  data.azurerm_sql_server.primary_sql_db_server.id)
}
