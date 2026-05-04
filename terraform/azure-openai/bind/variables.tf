variable "endpoint" { type = string }
variable "api_key" {
  type      = string
  sensitive = true
}
variable "api_version" { type = string }
variable "deployments" { type = string }
variable "ttl_expires_at" { type = string }
