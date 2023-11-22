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

resource "azurerm_resource_group" "rg" {
  name     = local.resource_group
  location = var.location
  tags     = var.labels
  count    = length(var.resource_group) == 0 ? 1 : 0

  lifecycle {
    prevent_destroy = true
  }
}

resource "azurerm_cosmosdb_account" "cosmosdb-account" {
  depends_on          = [azurerm_resource_group.rg]
  name                = var.instance_name
  location            = var.location
  resource_group_name = local.resource_group
  offer_type          = "Standard"
  kind                = "GlobalDocumentDB"

  consistency_policy {
    consistency_level       = var.consistency_level
    max_interval_in_seconds = var.max_interval_in_seconds
    max_staleness_prefix    = var.max_staleness_prefix
  }

  dynamic "geo_location" {
    for_each = var.failover_locations
    content {
      location          = geo_location.value
      failover_priority = index(var.failover_locations, geo_location.value)
    }
  }

  enable_automatic_failover         = var.enable_automatic_failover
  enable_multiple_write_locations   = var.enable_multiple_write_locations
  is_virtual_network_filter_enabled = local.enable_virtual_network_filter
  ip_range_filter                   = var.ip_range_filter
  tags                              = var.labels

  dynamic "virtual_network_rule" {
    for_each = var.authorized_network == "" ? [] : (var.authorized_network == "" ? [] : [1])
    content {
      id = var.authorized_network
    }
  }

}

resource "azurerm_cosmosdb_sql_database" "db" {
  name                = var.db_name
  resource_group_name = azurerm_cosmosdb_account.cosmosdb-account.resource_group_name
  account_name        = azurerm_cosmosdb_account.cosmosdb-account.name
  throughput          = var.request_units

  lifecycle {
    prevent_destroy = true
  }
}

resource "azurerm_private_endpoint" "private_endpoint" {
  name                = "${var.instance_name}-private_endpoint"
  location            = var.location
  resource_group_name = var.resource_group
  subnet_id           = var.private_endpoint_subnet_id
  tags                = var.labels
  count               = local.private_endpoint_enabled ? 1 : 0

  private_service_connection {
    name                           = "${var.instance_name}-private_service_connection"
    private_connection_resource_id = azurerm_cosmosdb_account.cosmosdb-account.id
    subresource_names              = ["SQL"]
    is_manual_connection           = false
  }

  dynamic "private_dns_zone_group" {
    for_each = length(var.private_dns_zone_ids) == 0 ? [] : [1]
    content {
      name                 = "${var.instance_name}-private_dns_zone_group"
      private_dns_zone_ids = var.private_dns_zone_ids
    }
  }
}