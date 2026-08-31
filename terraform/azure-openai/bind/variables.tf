variable "instance_name" { type = string }
variable "resource_tags_json" { type = string }
variable "endpoint" { type = string }
variable "api_key" {
  type      = string
  sensitive = true
}
variable "api_version" { type = string }
variable "deployments" { type = string }
variable "ttl_expires_at" { type = string }
variable "budget_enforcement_mode" { type = string }
variable "cf_context_json" { type = string }
variable "cf_originating_identity_json" { type = string }
variable "cf_app_guid" {
  type    = string
  default = ""
}
