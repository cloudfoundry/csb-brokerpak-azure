provider "postgresql" {
  host      = var.hostname
  port      = var.port
  username  = var.admin_username
  password  = var.admin_password
  superuser = false
  database  = var.db_name
  sslmode   = var.use_tls ? "require" : "disable"
}