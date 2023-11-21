output "name" { value = azurerm_postgresql_database.instance-db.name }
output "hostname" { value = azurerm_postgresql_server.instance.fqdn }
output "port" { value = 5432 }
output "username" { value = format("%s@%s", random_string.username.result, azurerm_postgresql_server.instance.name) }
output "password" {
  value     = random_password.password.result
  sensitive = true
}
output "use_tls" { value = var.use_tls }
output "status" { value = format("created db %s (id: %s) on server %s (id: %s) URL: https://portal.azure.com/#@%s/resource%s",
  azurerm_postgresql_database.instance-db.name,
  azurerm_postgresql_database.instance-db.id,
  azurerm_postgresql_server.instance.name,
  azurerm_postgresql_server.instance.id,
  var.azure_tenant_id,
  azurerm_postgresql_server.instance.id) }