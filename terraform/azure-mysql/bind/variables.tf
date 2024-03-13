variable "mysql_db_name" { type = string }
variable "mysql_hostname" { type = string }
variable "mysql_port" { type = number }
variable "admin_username" { type = string }
variable "admin_password" {
  type      = string
  sensitive = true
}
variable "use_tls" { type = bool }