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

resource "azurerm_resource_group" "azure-msyql" {
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

resource "random_string" "servername" {
  length  = 8
  special = false
}

resource "random_password" "password" {
  length           = 31
  override_special = "~_-."
  min_upper        = 2
  min_lower        = 2
  min_special      = 2
}

resource "random_string" "random" {
  length  = 8
  special = false
  upper   = false
}

resource "azurerm_mysql_server" "instance" {
  depends_on                       = [azurerm_resource_group.azure-msyql]
  name                             = lower(random_string.servername.result)
  location                         = var.location
  resource_group_name              = local.resource_group
  sku_name                         = local.sku_name
  storage_mb                       = var.storage_gb * 1024
  administrator_login              = random_string.username.result
  administrator_login_password     = random_password.password.result
  version                          = var.mysql_version
  ssl_enforcement_enabled          = var.use_tls
  ssl_minimal_tls_version_enforced = local.tls_version
  backup_retention_days            = var.backup_retention_days
  auto_grow_enabled                = true
  public_network_access_enabled    = local.private_endpoint_enabled ? false : true

  dynamic "threat_detection_policy" {
    for_each = var.enable_threat_detection_policy == null ? [] : (var.enable_threat_detection_policy ? [1] : [])
    content {
      enabled              = true
      email_addresses      = var.threat_detection_policy_emails == null ? [] : var.threat_detection_policy_emails
      email_account_admins = var.email_account_admins == null ? false : var.email_account_admins
    }
  }

  tags = var.labels

  lifecycle {
    prevent_destroy = true
  }
}

resource "azurerm_mysql_database" "instance-db" {
  name                = var.db_name
  resource_group_name = local.resource_group
  server_name         = azurerm_mysql_server.instance.name
  charset             = "utf8"
  collation           = "utf8_unicode_ci"

  lifecycle {
    prevent_destroy = true
  }
}

resource "azurerm_mysql_virtual_network_rule" "allow_subnet_id" {
  name                = format("subnetrule-%s", lower(random_string.servername.result))
  resource_group_name = local.resource_group
  server_name         = azurerm_mysql_server.instance.name
  subnet_id           = var.authorized_network
  count               = var.authorized_network != "default" ? 1 : 0
  depends_on          = [azurerm_mysql_database.instance-db]
}

resource "azurerm_mysql_virtual_network_rule" "allow_subnet_ids" {
  name                = format("subnetrule-%s-%s", lower(random_string.servername.result), count.index)
  resource_group_name = local.resource_group
  server_name         = azurerm_mysql_server.instance.name
  subnet_id           = var.authorized_networks[count.index]
  count               = length(var.authorized_networks)
  depends_on          = [azurerm_mysql_database.instance-db]
}

resource "azurerm_mysql_firewall_rule" "allow_azure" {
  name                = format("firewall-%s", lower(random_string.servername.result))
  resource_group_name = local.resource_group
  server_name         = azurerm_mysql_server.instance.name
  start_ip_address    = "0.0.0.0"
  end_ip_address      = "0.0.0.0"
  count               = var.authorized_network == "default" && local.private_endpoint_enabled == false ? 1 : 0
  depends_on          = [azurerm_mysql_database.instance-db]
}

resource "azurerm_mysql_firewall_rule" "allow_firewall" {
  name                = format("firewall_%s_%s", replace(var.instance_name, "-", "_"), count.index)
  resource_group_name = local.resource_group
  server_name         = azurerm_mysql_server.instance.name
  start_ip_address    = var.firewall_rules[count.index][0]
  end_ip_address      = var.firewall_rules[count.index][1]
  count               = length(var.firewall_rules)
  depends_on          = [azurerm_mysql_database.instance-db]
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
    private_connection_resource_id = azurerm_mysql_server.instance.id
    subresource_names              = ["mysqlServer"]
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