output "username" { value = random_string.username.result }
output "password" {
  value     = random_password.password.result
  sensitive = true
}
output "jdbcUrl" {
  value = format(
    "jdbc:sqlserver://%s:%d;database=%s;user=%s;password=%s;Encrypt=true;TrustServerCertificate=false;HostNameInCertificate=*.database.windows.net;loginTimeout=30",
    var.mssql_hostname,
    var.mssql_port,
    var.mssql_db_name,
    random_string.username.result,
    random_password.password.result,
  )
  sensitive = true
}
output "jdbcUrlForAuditingEnabled" {
  value = format(
    "jdbc:sqlserver://%s:%d;database=%s;user=%s;password=%s;Encrypt=true;TrustServerCertificate=false;HostNameInCertificate=*.database.windows.net;loginTimeout=30",
    var.mssql_hostname,
    var.mssql_port,
    var.mssql_db_name,
    random_string.username.result,
    random_password.password.result,
  )
  sensitive = true
}
output "uri" {
  value = format(
    "mssql://%s:%d/%s?encrypt=true&TrustServerCertificate=false&HostNameInCertificate=*.database.windows.net",
    var.mssql_hostname,
    var.mssql_port,
    var.mssql_db_name,
  )
  sensitive = true
}
output "databaseLogin" { value = random_string.username.result }
output "databaseLoginPassword" {
  value     = random_password.password.result
  sensitive = true
}