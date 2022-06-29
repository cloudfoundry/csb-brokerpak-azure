provider "csbsqlserver" {
  server   = var.mssql_hostname
  port     = var.mssql_port
  username = local.admin_username
  password = local.admin_password
  database = var.mssql_db_name
  encrypt  = "false" # Not ideal, but this matches what happened with the psqlcmd tool
}