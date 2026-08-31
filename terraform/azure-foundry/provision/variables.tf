variable "instance_name" { type = string }
variable "location" { type = string }
variable "resource_group" { type = string }
variable "sku_name" { type = string }
variable "deployment_sku_name" { type = string }
variable "ttl_hours" { type = number }
variable "budget_amount" { type = number }
variable "labels" { type = map(any) }
variable "foundry_hub_name" { type = string }
variable "foundry_project_name" { type = string }
variable "deployment_name" { type = string }
variable "model_name" { type = string }
variable "model_version" { type = string }
variable "model_capacity" { type = number }
variable "token_audience" { type = string }
variable "cf_context_json" { type = string }
variable "cf_originating_identity_json" { type = string }
variable "azure_tenant_id" { type = string }
variable "azure_subscription_id" { type = string }
variable "azure_client_id" { type = string }
variable "azure_client_secret" {
  type      = string
  sensitive = true
}

variable "budget_contact_email" {
  type    = string
  default = ""
}

variable "budget_webhook_url" {
  type    = string
  default = ""
}