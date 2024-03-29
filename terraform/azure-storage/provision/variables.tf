variable "storage_account_type" { type = string }
variable "tier" { type = string }
variable "replication_type" { type = string }
variable "location" { type = string }
variable "labels" { type = map(any) }
variable "resource_group" { type = string }
variable "azure_tenant_id" { type = string }
variable "azure_subscription_id" { type = string }
variable "azure_client_id" { type = string }
variable "azure_client_secret" {
  type      = string
  sensitive = true
}
variable "skip_provider_registration" { type = bool }
variable "authorized_networks" { type = list(string) }