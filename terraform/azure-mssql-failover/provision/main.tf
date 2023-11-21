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

resource "azurerm_resource_group" "azure-sql-fog" {
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
  length           = 64
  override_special = "~_-."
  min_upper        = 2
  min_lower        = 2
  min_special      = 2
}

resource "azurerm_sql_server" "primary_azure_sql_db_server" {
  depends_on                   = [azurerm_resource_group.azure-sql-fog]
  name                         = format("%s-primary", var.instance_name)
  resource_group_name          = local.resource_group
  location                     = var.location
  version                      = "12.0"
  administrator_login          = random_string.username.result
  administrator_login_password = random_password.password.result
  tags                         = var.labels

  lifecycle {
    prevent_destroy = true
  }
}


resource "azurerm_sql_server" "secondary_sql_db_server" {
  depends_on                   = [azurerm_resource_group.azure-sql-fog]
  name                         = format("%s-secondary", var.instance_name)
  resource_group_name          = local.resource_group
  location                     = var.failover_location != "default" ? var.location : local.default_pair[var.location]
  version                      = "12.0"
  administrator_login          = random_string.username.result
  administrator_login_password = random_password.password.result
  tags                         = var.labels

  lifecycle {
    prevent_destroy = true
  }
}

resource "azurerm_mssql_database" "azure_sql_db" {
  name                        = var.db_name
  server_id                   = azurerm_sql_server.primary_azure_sql_db_server.id
  sku_name                    = local.sku_name
  max_size_gb                 = var.max_storage_gb
  tags                        = var.labels
  min_capacity                = var.min_capacity
  auto_pause_delay_in_minutes = var.auto_pause_delay

  lifecycle {
    prevent_destroy = true
  }
}

resource "azurerm_mssql_database" "secondary_azure_sql_db" {
  name                        = var.db_name
  server_id                   = azurerm_sql_server.secondary_sql_db_server.id
  sku_name                    = local.sku_name
  tags                        = var.labels
  create_mode                 = "Secondary"
  creation_source_database_id = azurerm_mssql_database.azure_sql_db.id

  lifecycle {
    prevent_destroy = true
  }
}

resource "azurerm_sql_failover_group" "failover_group" {
  depends_on          = [azurerm_resource_group.azure-sql-fog]
  name                = var.instance_name
  resource_group_name = local.resource_group
  server_name         = azurerm_sql_server.primary_azure_sql_db_server.name
  databases           = [azurerm_mssql_database.azure_sql_db.id]
  partner_servers {
    id = azurerm_sql_server.secondary_sql_db_server.id
  }

  read_write_endpoint_failover_policy {
    mode          = var.read_write_endpoint_failover_policy
    grace_minutes = var.failover_grace_minutes
  }
}

resource "azurerm_sql_virtual_network_rule" "allow_subnet_id1" {
  name                = format("subnetrule1-%s", lower(var.instance_name))
  resource_group_name = local.resource_group
  server_name         = azurerm_sql_server.primary_azure_sql_db_server.name
  subnet_id           = var.authorized_network
  count               = var.authorized_network != "default" ? 1 : 0
}

resource "azurerm_sql_virtual_network_rule" "allow_subnet_id2" {
  name                = format("subnetrule2-%s", lower(var.instance_name))
  resource_group_name = local.resource_group
  server_name         = azurerm_sql_server.secondary_sql_db_server.name
  subnet_id           = var.authorized_network
  count               = var.authorized_network != "default" ? 1 : 0
}

resource "azurerm_sql_firewall_rule" "server1" {
  depends_on          = [azurerm_resource_group.azure-sql-fog]
  name                = format("firewallrule1-%s", lower(var.instance_name))
  resource_group_name = local.resource_group
  server_name         = azurerm_sql_server.primary_azure_sql_db_server.name
  start_ip_address    = "0.0.0.0"
  end_ip_address      = "0.0.0.0"
  count               = var.authorized_network == "default" ? 1 : 0
}

resource "azurerm_sql_firewall_rule" "server2" {
  depends_on          = [azurerm_resource_group.azure-sql-fog]
  name                = format("firewallrule2-%s", lower(var.instance_name))
  resource_group_name = local.resource_group
  server_name         = azurerm_sql_server.secondary_sql_db_server.name
  start_ip_address    = "0.0.0.0"
  end_ip_address      = "0.0.0.0"
  count               = var.authorized_network == "default" ? 1 : 0
}