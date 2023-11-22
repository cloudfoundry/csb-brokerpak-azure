output "name" { value = azurerm_mysql_database.instance-db.name }
output "hostname" { value = azurerm_mysql_server.instance.fqdn }
output "port" { value = 3306 }
output "username" { value = format("%s@%s", azurerm_mysql_server.instance.administrator_login, azurerm_mysql_server.instance.name) }
output "password" {
  value     = azurerm_mysql_server.instance.administrator_login_password
  sensitive = true
}
output "use_tls" { value = var.use_tls }
output "status" { value = format("created db %s (id: %s) on server %s (id: %s) URL: https://portal.azure.com/#@%s/resource%s",
  azurerm_mysql_database.instance-db.name,
  azurerm_mysql_database.instance-db.id,
  azurerm_mysql_server.instance.name,
  azurerm_mysql_server.instance.id,
  var.azure_tenant_id,
azurerm_mysql_server.instance.id) }