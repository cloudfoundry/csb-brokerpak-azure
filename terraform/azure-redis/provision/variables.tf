variable "resource_group" { type = string }
variable "azure_tenant_id" { type = string }
variable "azure_subscription_id" { type = string }
variable "azure_client_id" { type = string }
variable "azure_client_secret" {
  type      = string
  sensitive = true
}
variable "sku_name" { type = string }
variable "redis_version" { type = string }
variable "family" { type = string }
variable "capacity" { type = string }
variable "instance_name" { type = string }
variable "location" { type = string }
variable "labels" { type = map(any) }
variable "skip_provider_registration" { type = bool }
variable "tls_min_version" { type = string }
variable "maxmemory_policy" { type = string }
variable "firewall_rules" { type = list(list(string)) }
variable "subnet_id" { type = string }
variable "private_endpoint_subnet_id" { type = string }
variable "private_dns_zone_ids" { type = list(string) }