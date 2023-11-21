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

resource "random_string" "random" {
  length  = 8
  special = false
  upper   = false
}

resource "azurerm_resource_group" "azure-redis" {
  name     = local.resource_group
  location = var.location
  tags     = var.labels
  count    = length(var.resource_group) == 0 ? 1 : 0

  lifecycle {
    prevent_destroy = true
  }
}

resource "azurerm_redis_cache" "redis" {
  depends_on                    = [azurerm_resource_group.azure-redis]
  name                          = var.instance_name
  sku_name                      = var.sku_name
  redis_version                 = var.redis_version
  family                        = var.family
  capacity                      = var.capacity
  location                      = var.location
  resource_group_name           = local.resource_group
  minimum_tls_version           = length(var.tls_min_version) == 0 ? "1.2" : var.tls_min_version
  public_network_access_enabled = local.private_endpoint_enabled ? false : true
  tags                          = var.labels
  redis_configuration {
    maxmemory_policy = length(var.maxmemory_policy) == 0 ? "allkeys-lru" : var.maxmemory_policy
  }
  subnet_id = lower(var.sku_name) == "premium" && length(var.subnet_id) > 0 ? var.subnet_id : null

  lifecycle {
    prevent_destroy = true
  }
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
  count               = local.private_endpoint_enabled ? 1 : 0

  private_service_connection {
    name                           = "${random_string.random.result}-privateserviceconnection"
    private_connection_resource_id = azurerm_redis_cache.redis.id
    subresource_names              = ["redisCache"]
    is_manual_connection           = false
  }

  dynamic "private_dns_zone_group" {
    for_each = length(var.private_dns_zone_ids) == 0 ? [] : [1]
    content {
      name                 = "${random_string.random.result}-privatednszonegroup"
      private_dns_zone_ids = var.private_dns_zone_ids
    }
  }
}