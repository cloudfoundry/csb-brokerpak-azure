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

variable "fog_instance_name" { type = string }
variable "server_pair_name" { type = string }
variable "server_pairs" { type = map(any) }
variable "azure_tenant_id" { type = string }
variable "azure_subscription_id" { type = string }
variable "azure_client_id" { type = string }
variable "azure_client_secret" { type = string }

resource "null_resource" "run-failover" {

  triggers = {
    invoke = uuid()
    server_pair_name = var.server_pair_name
    fog_instance_name = var.fog_instance_name
    azure_subscription_id = var.azure_subscription_id
    azure_tenant_id      = var.azure_tenant_id
    azure_client_id       = var.azure_client_id
    azure_client_secret   = var.azure_client_secret
    secondary_resource_group = var.server_pairs[var.server_pair_name].secondary.resource_group
    secondary_server_name = var.server_pairs[var.server_pair_name].secondary.server_name
    primary_resource_group = var.server_pairs[var.server_pair_name].primary.resource_group
    primary_server_name = var.server_pairs[var.server_pair_name].primary.server_name
  }


  provisioner "local-exec" {
    command = format("sqlfailover %s %s %s",
      self.triggers.secondary_resource_group,
    self.triggers.secondary_server_name,
    self.triggers.fog_instance_name)
    environment = {
      ARM_SUBSCRIPTION_ID = self.triggers.azure_subscription_id
      ARM_TENANT_ID       = self.triggers.azure_tenant_id
      ARM_CLIENT_ID       = self.triggers.azure_client_id
      ARM_CLIENT_SECRET   = self.triggers.azure_client_secret
    }
  }

  provisioner "local-exec" {
    when = destroy
    command = format("sqlfailover %s %s %s",
    self.triggers.primary_resource_group,
    self.triggers.primary_server_name,
    self.triggers.fog_instance_name)
    environment = {
      ARM_SUBSCRIPTION_ID = self.triggers.azure_subscription_id
      ARM_TENANT_ID       = self.triggers.azure_tenant_id
      ARM_CLIENT_ID       = self.triggers.azure_client_id
      ARM_CLIENT_SECRET   = self.triggers.azure_client_secret
    }
  }
}
