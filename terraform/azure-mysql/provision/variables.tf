variable "instance_name" { type = string }
variable "resource_group" { type = string }
variable "azure_tenant_id" { type = string }
variable "azure_subscription_id" { type = string }
variable "azure_client_id" { type = string }
variable "azure_client_secret" {
  type      = string
  sensitive = true
}
variable "db_name" { type = string }
variable "mysql_version" { type = string }
variable "location" { type = string }
variable "labels" { type = map(any) }
variable "cores" { type = string }
variable "sku_name" { type = string }
variable "storage_gb" { type = string }
variable "authorized_network" { type = string }
variable "authorized_networks" { type = list(string) }
variable "use_tls" { type = bool }
variable "tls_min_version" { type = string }
variable "skip_provider_registration" { type = bool }
variable "backup_retention_days" { type = number }
variable "enable_threat_detection_policy" { type = bool }
variable "threat_detection_policy_emails" { type = list(string) }
variable "email_account_admins" { type = bool }
variable "firewall_rules" { type = list(list(string)) }
variable "private_endpoint_subnet_id" { type = string }
variable "private_dns_zone_ids" { type = list(string) }