output "username" { value = local.username }
output "password" {
  value     = random_password.password.result
  sensitive = true
}
output "uri" {
  value = format("%s://%s:%s@%s:%d/%s",
    "mysql",
    urlencode(local.username),
    random_password.password.result,
    var.mysql_hostname,
    var.mysql_port,
    var.mysql_db_name)
  sensitive = true
}
output "jdbcUrl" {
  value = format("jdbc:%s://%s:%s/%s?user=%s\u0026password=%s\u0026verifyServerCertificate=true\u0026useSSL=%v\u0026requireSSL=%v\u0026serverTimezone=GMT",
    "mysql",
    var.mysql_hostname,
    var.mysql_port,
    var.mysql_db_name,
    urlencode(local.username),
    random_password.password.result,
    var.use_tls,
    var.use_tls)
  sensitive = true
}