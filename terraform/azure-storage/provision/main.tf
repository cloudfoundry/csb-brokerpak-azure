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

resource "random_string" "account_name" {
  length  = 24
  special = false
  upper   = false
}

resource "azurerm_resource_group" "azure-storage" {
  name     = local.resource_group
  location = var.location
  tags     = var.labels
  count    = length(var.resource_group) == 0 ? 1 : 0

  lifecycle {
    prevent_destroy = true
  }
}

resource "azurerm_storage_account" "account" {
  depends_on               = [azurerm_resource_group.azure-storage]
  name                     = random_string.account_name.result
  resource_group_name      = local.resource_group
  location                 = var.location
  account_tier             = var.tier
  account_replication_type = var.replication_type
  account_kind             = var.storage_account_type

  tags = var.labels

  lifecycle {
    prevent_destroy = true
  }
}

resource "azurerm_storage_account_network_rules" "account_network_rule" {
  count = length(var.authorized_networks) != 0 ? 1 : 0

  storage_account_id = azurerm_storage_account.account.id

  default_action             = "Deny"
  virtual_network_subnet_ids = var.authorized_networks[*]
}