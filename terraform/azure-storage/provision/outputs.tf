output "primary_access_key" {
  value     = azurerm_storage_account.account.primary_access_key
  sensitive = true
}
output "secondary_access_key" {
  value     = azurerm_storage_account.account.secondary_access_key
  sensitive = true
}
output "storage_account_name" { value = azurerm_storage_account.account.name }
output "status" { value = format("created storage account %s (id: %s) URL:  https://portal.azure.com/#@%s/resource%s",
  azurerm_storage_account.account.name,
  azurerm_storage_account.account.id,
  var.azure_tenant_id,
azurerm_storage_account.account.id) }