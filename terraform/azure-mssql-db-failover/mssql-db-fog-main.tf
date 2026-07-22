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

resource "azurerm_mssql_database" "primary_db" {
  name           = var.db_name
  server_id      = data.azurerm_mssql_server.primary_sql_db_server.id
  sku_name       = local.sku_name
  max_size_gb    = var.max_storage_gb
  tags           = var.labels
  zone_redundant = var.primary_zone_redundant
  count          = var.existing ? 0 : 1
  short_term_retention_policy {
    retention_days = var.short_term_retention_days
  }
  dynamic "long_term_retention_policy" {
    for_each = local.valid_ltr_policy ? [1] : []
    content {
      weekly_retention  = var.ltr_weekly_retention
      monthly_retention = var.ltr_monthly_retention
      yearly_retention  = var.ltr_yearly_retention
      week_of_year      = var.ltr_week_of_year
    }
  }
}

resource "azurerm_mssql_database" "secondary_db" {
  name                        = var.db_name
  server_id                   = data.azurerm_mssql_server.secondary_sql_db_server.id
  sku_name                    = local.sku_name
  tags                        = var.labels
  zone_redundant              = var.secondary_zone_redundant
  create_mode                 = "Secondary"
  creation_source_database_id = azurerm_mssql_database.primary_db[count.index].id
  count                       = var.existing ? 0 : 1
}

resource "azurerm_mssql_failover_group" "failover_group" {
  name      = var.instance_name
  server_id = data.azurerm_mssql_server.primary_sql_db_server.id
  databases = [
    var.existing ? data.azurerm_mssql_database.existing_primary_db[0].id : azurerm_mssql_database.primary_db[0].id
  ]
  tags = var.labels

  partner_server {
    id = data.azurerm_mssql_server.secondary_sql_db_server.id
  }

  read_write_endpoint_failover_policy {
    mode          = var.read_write_endpoint_failover_policy
    grace_minutes = var.read_write_endpoint_failover_policy == "Automatic" ? var.failover_grace_minutes : null
  }
}
