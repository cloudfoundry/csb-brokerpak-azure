variable "instance_name" { type = string }
variable "location" { type = string }
variable "resource_group" { type = string }
variable "sku_name" { type = string }
variable "ttl_hours" { type = number }
variable "gpt35_capacity" { type = number }
variable "embedding_capacity" { type = number }
variable "budget_amount" { type = number }
variable "labels" { type = map(any) }

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
