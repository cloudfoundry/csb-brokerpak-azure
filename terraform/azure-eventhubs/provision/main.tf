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

resource "azurerm_eventhub_namespace" "rg-namespace" {
  depends_on           = [azurerm_resource_group.rg]
  name                 = var.instance_name
  location             = var.location
  resource_group_name  = local.resource_group
  sku                  = var.sku
  capacity             = 1
  auto_inflate_enabled = var.auto_inflate_enabled
  tags                 = local.tags
}

resource "azurerm_eventhub" "eventhub" {
  name                = var.instance_name
  namespace_name      = azurerm_eventhub_namespace.rg-namespace.name
  resource_group_name = local.resource_group
  partition_count     = var.partition_count
  message_retention   = var.message_retention

  lifecycle {
    prevent_destroy = true
  }
}
