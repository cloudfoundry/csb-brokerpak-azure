# Copyright 2020 Pivotal Software, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

variable resource_group { type = string }
variable azure_tenant_id { type = string }
variable azure_subscription_id { type = string }
variable azure_client_id { type = string }
variable azure_client_secret { type = string }
variable sku_name { type = string }
variable family { type = string }
variable capacity { type = string }
variable instance_name { type = string }
variable location { type = string }
variable labels { type = map }
variable skip_provider_registration { type = bool }
variable tls_min_version { type = string }
variable maxmemory_policy { type = string }
variable firewall_rules { type = list(list(string)) }
variable subnet_id { type = string }
variable private_endpoint_subnet_id { type = string }
variable private_dns_zone_ids { type = list(string) }

provider "azurerm" {
  version = ">= 2.33.0"
  features {}

  subscription_id = var.azure_subscription_id
  client_id       = var.azure_client_id
  client_secret   = var.azure_client_secret
  tenant_id       = var.azure_tenant_id  

  skip_provider_registration = var.skip_provider_registration
}

resource "random_string" "random" {
  length = 8
  special = false
  upper = false
}

locals {
  resource_group = length(var.resource_group) == 0 ? format("rg-%s", var.instance_name) : var.resource_group
  private_endpoint_enabled = var.private_endpoint_subnet_id == null ? false : length(var.private_endpoint_subnet_id) > 0 ? true : false
}

resource "azurerm_resource_group" "azure-redis" {
  name     = local.resource_group
  location = var.location
  tags     = var.labels
  count    = length(var.resource_group) == 0 ? 1 : 0
}

resource "azurerm_redis_cache" "redis" {
  depends_on  = [ azurerm_resource_group.azure-redis ]  
  name                = var.instance_name
  sku_name            = var.sku_name
  family              = var.family
  capacity            = var.capacity
  location            = var.location
  resource_group_name = local.resource_group
  minimum_tls_version = length(var.tls_min_version) == 0 ? "1.2" : var.tls_min_version
  public_network_access_enabled = local.private_endpoint_enabled ? false : true
  tags                = var.labels
  redis_configuration {
    maxmemory_policy   = length(var.maxmemory_policy) == 0 ? "allkeys-lru" : var.maxmemory_policy
  }
  subnet_id = lower(var.sku_name) == "premium" && length(var.subnet_id) > 0 ? var.subnet_id : null
}

resource "azurerm_redis_firewall_rule" "allow_azure" {
  name                = format("firewall_%s_%s", replace(var.instance_name, "-", "_"), count.index)
  resource_group_name = local.resource_group
  redis_cache_name    = azurerm_redis_cache.redis.name
  start_ip            = var.firewall_rules[count.index][0]
  end_ip              = var.firewall_rules[count.index][1]

  count = length(var.firewall_rules)
}    

resource "azurerm_private_endpoint" "private_endpoint" {
  name                = "${random_string.random.result}-privateendpoint"
  location            = var.location
  resource_group_name = var.resource_group
  subnet_id           = var.private_endpoint_subnet_id
  tags                = var.labels
  count = local.private_endpoint_enabled ? 1 : 0

  private_service_connection {
    name                           = "${random_string.random.result}-privateserviceconnection"
    private_connection_resource_id = azurerm_redis_cache.redis.id
    subresource_names              = [ "redisCache" ]
    is_manual_connection           = false
  }

  dynamic "private_dns_zone_group" {
    for_each = length(var.private_dns_zone_ids) == 0 ? [] : [1]
    content {
      name = "${random_string.random.result}-privatednszonegroup"
      private_dns_zone_ids = var.private_dns_zone_ids
    }
  }
}

output name { value = azurerm_redis_cache.redis.name }
output host { value = azurerm_redis_cache.redis.hostname }
# output port { value = azurerm_redis_cache.redis.port }
output password { value = azurerm_redis_cache.redis.primary_access_key }
output tls_port { value = azurerm_redis_cache.redis.ssl_port }
output status { value = format("created cache %s (id: %s) URL: URL: https://portal.azure.com/#@%s/resource%s",
                               azurerm_redis_cache.redis.name,
                               azurerm_redis_cache.redis.id,
                               var.azure_tenant_id,
                               azurerm_redis_cache.redis.id)}