variable "instance_name" { type = string }
variable "azure_tenant_id" { type = string }
variable "azure_subscription_id" { type = string }
variable "azure_client_id" { type = string }
variable "azure_client_secret" {
  type      = string
  sensitive = true
}
variable "resource_group" { type = string }
variable "db_name" { type = string }
variable "location" { type = string }
variable "failover_location" { type = string }
variable "labels" { type = map(any) }
variable "sku_name" { type = string }
variable "cores" { type = number }
variable "max_storage_gb" { type = number }
variable "min_capacity" { type = number }
variable "auto_pause_delay" { type = number }
variable "authorized_network" { type = string }
variable "skip_provider_registration" { type = bool }
variable "read_write_endpoint_failover_policy" { type = string }
variable "failover_grace_minutes" { type = number }