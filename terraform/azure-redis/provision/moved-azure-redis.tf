moved {
  from = module.instance.random_string.random
  to   = random_string.random
}
moved {
  from = module.instance.azurerm_resource_group.azure-redis
  to   = azurerm_resource_group.azure-redis
}
moved {
  from = module.instance.azurerm_redis_cache.redis
  to   = azurerm_redis_cache.redis
}
moved {
  from = module.instance.azurerm_redis_firewall_rule.allow_azure
  to   = azurerm_redis_firewall_rule.allow_azure
}
moved {
  from = module.instance.azurerm_private_endpoint.private_endpoint
  to   = azurerm_private_endpoint.private_endpoint
}
