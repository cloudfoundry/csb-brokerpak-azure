variable "db_name" { type = string }
variable "hostname" { type = string }
variable "port" { type = number }
variable "admin_username" { type = string }
variable "admin_password" {
  type      = string
  sensitive = true
}