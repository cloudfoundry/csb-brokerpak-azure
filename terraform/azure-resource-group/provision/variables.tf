variable "instance_name" { type = string }
variable "azure_tenant_id" { type = string }
variable "azure_subscription_id" { type = string }
variable "azure_client_id" { type = string }
variable "azure_client_secret" {
  type      = string
  sensitive = true
}
variable "location" { type = string }
variable "labels" { type = map(any) }
variable "skip_provider_registration" { type = bool }