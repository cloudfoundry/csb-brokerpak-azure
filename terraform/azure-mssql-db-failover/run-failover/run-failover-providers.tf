variable "azure_tenant_id" {
  type      = string
  sensitive = true
}
variable "azure_subscription_id" {
  type      = string
  sensitive = true
}
variable "azure_client_id" {
  type      = string
  sensitive = true
}
variable "azure_client_secret" {
  type      = string
  sensitive = true
}

provider "csbmssqldbrunfailover" {
  azure_tenant_id       = var.azure_tenant_id
  azure_client_id       = var.azure_client_id
  azure_client_secret   = var.azure_client_secret
  azure_subscription_id = var.azure_subscription_id
}
