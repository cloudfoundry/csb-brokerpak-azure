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
variable "location" { type = string }
variable "sku" { type = string }
variable "auto_inflate_enabled" { type = bool }
variable "partition_count" { type = number }
variable "message_retention" { type = number }
variable "labels" { type = map(any) }
