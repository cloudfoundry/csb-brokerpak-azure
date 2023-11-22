provider "mysql" {
  endpoint = format("%s:%d", var.mysql_hostname, var.mysql_port)
  username = var.admin_username
  password = var.admin_password
  tls      = var.use_tls
}