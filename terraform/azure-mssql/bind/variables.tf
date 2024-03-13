variable "mssql_db_name" { type = string }
variable "mssql_hostname" { type = string }
variable "mssql_port" { type = number }
variable "admin_username" { type = string }
variable "admin_password" {
  type      = string
  sensitive = true
}