output "name" { value = azurerm_redis_cache.redis.name }
output "host" { value = azurerm_redis_cache.redis.hostname }
# output port { value = azurerm_redis_cache.redis.port }
output "password" {
  value     = azurerm_redis_cache.redis.primary_access_key
  sensitive = true
}
output "tls_port" { value = azurerm_redis_cache.redis.ssl_port }
output "status" { value = format("created cache %s (id: %s) URL: https://portal.azure.com/#@%s/resource%s",
  azurerm_redis_cache.redis.name,
  azurerm_redis_cache.redis.id,
  var.azure_tenant_id,
azurerm_redis_cache.redis.id) }