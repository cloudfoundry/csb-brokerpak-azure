variable "instance_name" { type = string }
variable "implementation_state" { type = string }
variable "endpoint" { type = string }
variable "resource_tags_json" { type = string }
variable "api_key" {
  type      = string
  sensitive = true
}
variable "api_version" { type = string }
variable "deployments" { type = string }
variable "foundry_hub_name" { type = string }
variable "foundry_project_name" { type = string }
variable "deployment_name" { type = string }
variable "model_name" { type = string }
variable "model_version" { type = string }
variable "model_capacity" { type = number }
variable "resource_group" { type = string }
variable "location" { type = string }
variable "azure_tenant_id" { type = string }
variable "azure_subscription_id" { type = string }
variable "budget_enforcement_mode" { type = string }
variable "token_audience" { type = string }
variable "ttl_expires_at" { type = string }
variable "cf_context_json" { type = string }
variable "cf_originating_identity_json" { type = string }
variable "cf_app_guid" {
  type    = string
  default = ""
}