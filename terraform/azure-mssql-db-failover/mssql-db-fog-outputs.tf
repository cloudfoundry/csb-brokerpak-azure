# locals needed as Terraform >= 0.13 evaluates the output variables on TF import. As a result we attempt to access a resource which has not yet been
# instantiated and get an error. This check is to stop failures on the import step run for the subsume plan.
locals {
  primary_db_name = (length(azurerm_mssql_database.primary_db) > 0 ? azurerm_mssql_database.primary_db[0].name : "")
  primary_db_id   = (length(azurerm_mssql_database.primary_db) > 0 ? azurerm_mssql_database.primary_db[0].id : "")
  fog_name        = (length(azurerm_mssql_failover_group.failover_group) > 0 ? azurerm_mssql_failover_group.failover_group[0].name : "")
  fog_id          = (length(azurerm_mssql_failover_group.failover_group) > 0 ? azurerm_mssql_failover_group.failover_group[0].id : "")
}

output "sqldbName" { value = var.existing ? var.db_name : local.primary_db_name }
output "sqlServerName" { value = var.existing ? var.instance_name : local.fog_name }
output "sqlServerFullyQualifiedDomainName" { value = format("%s.database.windows.net", var.existing ? var.instance_name : local.fog_name) }
output "hostname" { value = format("%s.database.windows.net", var.existing ? var.instance_name : local.fog_name) }
output "port" { value = 1433 }
output "name" { value = var.existing ? var.db_name : local.primary_db_name }
output "username" {
  value     = var.server_credential_pairs[var.server_pair].admin_username
  sensitive = true
}
output "password" {
  value     = var.server_credential_pairs[var.server_pair].admin_password
  sensitive = true
}
output "server_pair" { value = var.server_pair }
output "status" {
  value = var.existing ? format("connected to existing failover group - primary server %s (id: %s) secondary server %s (%s) URL: https://portal.azure.com/#@%s/resource%s/failoverGroup",
    data.azurerm_mssql_server.primary_sql_db_server.name, data.azurerm_mssql_server.primary_sql_db_server.id,
    data.azurerm_mssql_server.secondary_sql_db_server.name, data.azurerm_mssql_server.secondary_sql_db_server.id,
    var.azure_tenant_id,
    data.azurerm_mssql_server.primary_sql_db_server.id) : format("created failover group %s (id: %s), primary db %s (id: %s) on server %s (id: %s), secondary db %s (id: %s/databases/%s) on server %s (id: %s) URL: https://portal.azure.com/#@%s/resource%s/failoverGroup",
    local.fog_name, local.fog_id,
    local.fog_name, local.primary_db_id,
    data.azurerm_mssql_server.primary_sql_db_server.name, data.azurerm_mssql_server.primary_sql_db_server.id,
    local.primary_db_name, data.azurerm_mssql_server.secondary_sql_db_server.id, local.primary_db_name,
    data.azurerm_mssql_server.secondary_sql_db_server.name, data.azurerm_mssql_server.secondary_sql_db_server.id,
    var.azure_tenant_id,
    data.azurerm_mssql_server.primary_sql_db_server.id,
  )
  sensitive = true
}
