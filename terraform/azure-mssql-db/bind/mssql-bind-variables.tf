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
variable "server" { type = string }
variable "server_credentials" { 
  type = map(any)
  sensitive = true
}
