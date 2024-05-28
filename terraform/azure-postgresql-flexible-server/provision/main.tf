# Copyright 2024 Broadcom Inc.
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

resource "azurerm_resource_group" "azure-postgres" {
  count    = length(var.resource_group) == 0 ? 1 : 0

  name     = local.resource_group
  location = var.location
  tags     = var.labels

  lifecycle {
    prevent_destroy = true
  }
}

resource "random_string" "username" {
  length  = 16
  special = false
  numeric  = false
}

resource "random_password" "password" {
  length           = 64
  override_special = "~_-."
  min_upper        = 2
  min_lower        = 2
  min_special      = 2
}

resource "azurerm_postgresql_flexible_server" "instance" {
  depends_on                   = [azurerm_resource_group.azure-postgres]

  name                         = var.instance_name
  resource_group_name          = local.resource_group
  location                     = var.location
  version                      = var.postgres_version
  sku_name                     = var.sku_name
  storage_mb                   = var.storage_gb * 1024
  administrator_login          = random_string.username.result
  administrator_password       = random_password.password.result
  tags                         = var.labels

  delegated_subnet_id          = var.delegated_subnet_id
  private_dns_zone_id          = var.delegated_subnet_id != null ? var.private_dns_zone_id : null

  lifecycle {
    prevent_destroy = true
  }
}

resource "azurerm_postgresql_flexible_server_database" "instance-db" {
  name                = var.db_name
  server_id           = azurerm_postgresql_flexible_server.instance.id
  charset             = "UTF8"
  collation           = "en_US.utf8"

  lifecycle {
    prevent_destroy = true
  }
}

resource "azurerm_postgresql_flexible_server_firewall_rule" "allow_azure" {
  count               = var.allow_access_from_azure_services ? 1 : 0

  name                = "allow-access-from-azure-services"
  server_id           = azurerm_postgresql_flexible_server.instance.id
  start_ip_address    = "0.0.0.0"
  end_ip_address      = "0.0.0.0"
}

resource "azurerm_private_endpoint" "private_endpoint" {
  count               = length(var.private_endpoint_subnet_id) != 0 ? 1 : 0

  name                = "${var.instance_name}-private_endpoint"
  location            = var.location
  resource_group_name = var.resource_group
  subnet_id           = var.private_endpoint_subnet_id
  tags                = var.labels

  private_service_connection {
    name                           = "${var.instance_name}-private_service_connection"
    private_connection_resource_id = azurerm_postgresql_flexible_server.instance.id
    subresource_names              = ["postgresqlServer"]
    is_manual_connection           = false
  }

  private_dns_zone_group {
    name                 = "${var.instance_name}-private_dns_zone_group"
    private_dns_zone_ids = [var.private_dns_zone_id]
  }
}