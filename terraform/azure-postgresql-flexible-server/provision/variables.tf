variable "azure_tenant_id" { type = string }
variable "azure_subscription_id" { type = string }
variable "azure_client_id" { type = string }
variable "azure_client_secret" {
  type      = string
  sensitive = true
}
variable "skip_provider_registration" { type = bool }
variable "labels" { type = map(any) }
variable "location" { type = string }
variable "resource_group" { type = string }
variable "postgres_version" { type = string }
variable "storage_gb" { type = number }
variable "sku_name" { type = string }
variable "instance_name" { type = string }
variable "db_name" { type = string }
variable "allow_access_from_azure_services" { type = bool }
variable "delegated_subnet_id" { type = string }
variable "private_dns_zone_id" { type = string }
variable "private_endpoint_subnet_id" { type = string }
