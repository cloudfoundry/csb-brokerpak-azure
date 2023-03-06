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

variable "cores" { type = number }
variable "instance_name" { type = string }
variable "db_name" { type = string }
variable "location" { type = string }
variable "labels" { type = map(any) }
variable "storage_gb" { type = number }
variable "resource_group" { type = string }
variable "azure_tenant_id" { type = string }
variable "azure_subscription_id" { type = string }
variable "azure_client_id" { type = string }
variable "azure_client_secret" { type = string }
variable "postgres_version" { type = string }
variable "sku_name" { type = string }
variable "authorized_network" { type = string }
variable "use_tls" { type = bool }
variable "skip_provider_registration" { type = bool }

terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = ">=2.33.0"
    }
    random = {
      source  = "hashicorp/random"
      version = ">=3.3.1"
    }
  }
}

provider "azurerm" {
  features {}

  subscription_id = var.azure_subscription_id
  client_id       = var.azure_client_id
  client_secret   = var.azure_client_secret
  tenant_id       = var.azure_tenant_id

  skip_provider_registration = var.skip_provider_registration
}

locals {
  instance_types = {
    1  = "GP_Gen5_1"
    2  = "GP_Gen5_2"
    4  = "GP_Gen5_4"
    8  = "GP_Gen5_8"
    16 = "GP_Gen5_16"
    32 = "GP_Gen5_32"
    64 = "GP_Gen5_64"
  }
  resource_group = length(var.resource_group) == 0 ? format("rg-%s", var.instance_name) : var.resource_group
  sku_name       = length(var.sku_name) == 0 ? local.instance_types[var.cores] : var.sku_name
}

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

output "name" { value = azurerm_postgresql_database.instance-db.name }
output "hostname" { value = azurerm_postgresql_server.instance.fqdn }
output "port" { value = 5432 }
output "username" { value = format("%s@%s", random_string.username.result, azurerm_postgresql_server.instance.name) }
output "password" {
  value     = random_password.password.result
  sensitive = true
}
output "use_tls" { value = var.use_tls }
output "status" { value = format("created db %s (id: %s) on server %s (id: %s) URL: https://portal.azure.com/#@%s/resource%s",
  azurerm_postgresql_database.instance-db.name,
  azurerm_postgresql_database.instance-db.id,
  azurerm_postgresql_server.instance.name,
  azurerm_postgresql_server.instance.id,
  var.azure_tenant_id,
azurerm_postgresql_server.instance.id) }