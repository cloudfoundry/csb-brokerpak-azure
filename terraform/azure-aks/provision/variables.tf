variable "instance_name" { type = string }
variable "location" { type = string }
variable "resource_group" { type = string }
variable "ttl_hours" { type = number }
variable "node_count" { type = number }
variable "node_vm_size" { type = string }
variable "private_cluster_enabled" { type = bool }
variable "authorized_api_server_cidrs" { type = list(string) }
variable "vnet_address_space" { type = list(string) }
variable "subnet_address_prefixes" { type = list(string) }
variable "service_cidr" { type = string }
variable "dns_service_ip" { type = string }
variable "pod_cidr" { type = string }
variable "labels" { type = map(any) }
variable "cf_context_json" { type = string }
variable "cf_originating_identity_json" { type = string }

variable "azure_tenant_id" { type = string }
variable "azure_subscription_id" { type = string }
variable "azure_client_id" { type = string }
variable "azure_client_secret" {
  type      = string
  sensitive = true
}
