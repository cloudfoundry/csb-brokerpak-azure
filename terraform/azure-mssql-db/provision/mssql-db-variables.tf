variable "azure_subscription_id" { type = string }
variable "azure_client_id" { type = string }
variable "azure_client_secret" {
  type      = string
  sensitive = true
}
variable "azure_tenant_id" { type = string }
variable "skip_provider_registration" { type = bool }
variable "db_name" { type = string }
variable "server" { type = string }
variable "server_credentials" {
  type      = map(any)
  sensitive = true
}
variable "labels" { type = map(any) }
variable "sku_name" { type = string }
variable "cores" { type = number }
variable "max_storage_gb" { type = number }
variable "short_term_retention_days" { type = number }
variable "ltr_weekly_retention" { type = string }
variable "ltr_monthly_retention" { type = string }
variable "ltr_yearly_retention" { type = string }
variable "ltr_week_of_year" { type = number }
