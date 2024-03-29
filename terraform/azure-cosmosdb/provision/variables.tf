variable "azure_tenant_id" { type = string }
variable "azure_subscription_id" { type = string }
variable "azure_client_id" { type = string }
variable "azure_client_secret" {
  type      = string
  sensitive = true
}
variable "skip_provider_registration" { type = bool }
variable "instance_name" { type = string }
variable "resource_group" { type = string }
variable "db_name" { type = string }
variable "failover_locations" { type = list(string) }
variable "location" { type = string }
variable "ip_range_filter" { type = string }
variable "request_units" { type = number }
variable "enable_automatic_failover" { type = bool }
variable "enable_multiple_write_locations" { type = bool }
variable "consistency_level" { type = string }
variable "max_interval_in_seconds" { type = number }
variable "max_staleness_prefix" { type = number }
variable "labels" { type = map(any) }
variable "authorized_network" { type = string }
variable "private_endpoint_subnet_id" { type = string }
variable "private_dns_zone_ids" { type = list(string) }