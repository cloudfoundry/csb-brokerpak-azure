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

resource "postgresql_role" "new_user" {
  name                = random_string.username.result
  login               = true
  password            = random_password.password.result
  skip_reassign_owned = true
  skip_drop_role      = true
}

resource "postgresql_grant" "all_access" {
  depends_on  = [postgresql_role.new_user]
  database    = var.db_name
  role        = random_string.username.result
  object_type = "database"
  privileges  = ["ALL"]
}

resource "postgresql_grant" "table_access" {
  depends_on  = [postgresql_role.new_user]
  database    = var.db_name
  role        = postgresql_role.new_user.name
  schema      = "public"
  object_type = "table"
  privileges  = ["ALL"]
}