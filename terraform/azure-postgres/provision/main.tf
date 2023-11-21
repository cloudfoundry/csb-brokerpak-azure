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

resource "azurerm_resource_group" "azure-postgres" {
  name     = local.resource_group
  location = var.location
  tags     = var.labels
  count    = length(var.resource_group) == 0 ? 1 : 0

  lifecycle {
    prevent_destroy = true
  }
}

resource "random_string" "username" {
  length  = 16
  special = false
  number  = false
}

resource "random_password" "password" {
  length           = 31
  override_special = "~_-."
  min_upper        = 2
  min_lower        = 2
  min_special      = 2
}

resource "azurerm_postgresql_server" "instance" {
  depends_on                   = [azurerm_resource_group.azure-postgres]
  name                         = var.instance_name
  location                     = var.location
  resource_group_name          = local.resource_group
  sku_name                     = local.sku_name
  storage_mb                   = var.storage_gb * 1024
  administrator_login          = random_string.username.result
  administrator_login_password = random_password.password.result
  version                      = var.postgres_version
  ssl_enforcement_enabled      = var.use_tls
  tags                         = var.labels

  lifecycle {
    prevent_destroy = true
  }
}

resource "azurerm_postgresql_database" "instance-db" {
  name                = var.db_name
  resource_group_name = local.resource_group
  server_name         = azurerm_postgresql_server.instance.name
  charset             = "UTF8"
  collation           = "en-US"

  lifecycle {
    prevent_destroy = true
  }
}

resource "azurerm_postgresql_virtual_network_rule" "allow_subnet_id" {
  name                = format("snr-%s", var.instance_name)
  resource_group_name = local.resource_group
  server_name         = azurerm_postgresql_server.instance.name
  subnet_id           = var.authorized_network
  count               = var.authorized_network != "default" ? 1 : 0
}

resource "azurerm_postgresql_firewall_rule" "allow_azure" {
  name                = format("f-%s", var.instance_name)
  resource_group_name = local.resource_group
  server_name         = azurerm_postgresql_server.instance.name
  start_ip_address    = "0.0.0.0"
  end_ip_address      = "0.0.0.0"
  count               = var.authorized_network == "default" ? 1 : 0
}