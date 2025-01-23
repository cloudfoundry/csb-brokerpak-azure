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
  name        = var.db_name
  server_id   = data.azurerm_mssql_server.primary_sql_db_server.id
  sku_name    = local.sku_name
  max_size_gb = var.max_storage_gb
  tags        = var.labels
  count       = var.existing ? 0 : 1
  short_term_retention_policy {
    retention_days = var.short_term_retention_days
  }
  long_term_retention_policy {
    weekly_retention  = var.ltr_weekly_retention
    monthly_retention = var.ltr_monthly_retention
    yearly_retention  = var.ltr_yearly_retention
    week_of_year      = var.ltr_week_of_year
  }
}

resource "azurerm_mssql_database" "secondary_db" {
  name                        = var.db_name
  server_id                   = data.azurerm_mssql_server.secondary_sql_db_server.id
  sku_name                    = local.sku_name
  tags                        = var.labels
  create_mode                 = "Secondary"
  creation_source_database_id = azurerm_mssql_database.primary_db[count.index].id
  count                       = var.existing ? 0 : 1
}

resource "azurerm_mssql_failover_group" "failover_group" {
  name      = var.instance_name
  server_id = data.azurerm_mssql_server.primary_sql_db_server.id
  databases = [azurerm_mssql_database.primary_db[count.index].id]
  partner_server {
    id = data.azurerm_mssql_server.secondary_sql_db_server.id
  }

  read_write_endpoint_failover_policy {
    mode          = var.read_write_endpoint_failover_policy
    grace_minutes = var.failover_grace_minutes
  }

  depends_on = [azurerm_mssql_database.secondary_db]

  count = var.existing ? 0 : 1
}
