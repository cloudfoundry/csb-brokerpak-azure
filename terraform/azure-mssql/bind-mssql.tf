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

variable "mssql_db_name" { type = string }
variable "mssql_hostname" { type = string }
variable "mssql_port" { type = number }
variable "admin_username" { type = string }
variable "admin_password" {
  type = string
  sensitive = true
}

terraform {
  required_providers {
    random = {
      source  = "hashicorp/random"
      version = ">=3.3.1"
    }
    csbsqlserver = {
      source  = "cloud-service-broker/csbsqlserver"
      version = "1.0.0"
    }
  }
}

provider "csbsqlserver" {
  server   = var.mssql_hostname
  port     = var.mssql_port
  username = var.admin_username
  password = var.admin_password
  database = var.mssql_db_name
  encrypt  = "false" # Not ideal, but this matches what happened with the psqlcmd tool
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

resource "csbsqlserver_binding" "binding" {
  username = random_string.username.result
  password = random_password.password.result
  roles    = ["db_ddladmin", "db_datareader", "db_datawriter", "db_accessadmin"]
}

output "username" { value = random_string.username.result }
output "password" {
  value     = random_password.password.result
  sensitive = true
}
output "jdbcUrl" {
  value = format(
    "jdbc:sqlserver://%s:%d;database=%s;user=%s;password=%s;Encrypt=true;TrustServerCertificate=false;HostNameInCertificate=*.database.windows.net;loginTimeout=30",
    var.mssql_hostname,
    var.mssql_port,
    var.mssql_db_name,
    random_string.username.result,
    random_password.password.result,
  )
  sensitive = true
}
output "jdbcUrlForAuditingEnabled" {
  value = format(
    "jdbc:sqlserver://%s:%d;database=%s;user=%s;password=%s;Encrypt=true;TrustServerCertificate=false;HostNameInCertificate=*.database.windows.net;loginTimeout=30",
    var.mssql_hostname,
    var.mssql_port,
    var.mssql_db_name,
    random_string.username.result,
    random_password.password.result,
  )
  sensitive = true
}
output "uri" {
  value = format(
    "mssql://%s:%d/%s?encrypt=true&TrustServerCertificate=false&HostNameInCertificate=*.database.windows.net",
    var.mssql_hostname,
    var.mssql_port,
    var.mssql_db_name,
  )
  sensitive = true
}
output "databaseLogin" { value = random_string.username.result }
output "databaseLoginPassword" {
  value     = random_password.password.result
  sensitive = true
}
