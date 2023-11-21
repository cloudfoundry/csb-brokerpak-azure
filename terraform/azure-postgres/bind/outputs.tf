output "username" { value = local.username }
output "password" {
  value     = random_password.password.result
  sensitive = true
}
output "uri" {
  value = format("%s://%s:%s@%s:%d/%s",
    "postgresql",
    urlencode(local.username),
    random_password.password.result,
    var.hostname,
    var.port,
    var.db_name)
  sensitive = true
}
output "jdbcUrl" {
  value = format("jdbc:%s://%s:%s/%s?user=%s\u0026password=%s\u0026verifyServerCertificate=true\u0026useSSL=%v\u0026requireSSL=false\u0026serverTimezone=GMT",
    "postgresql",
    var.hostname,
    var.port,
    var.db_name,
    urlencode(local.username),
    random_password.password.result,
    var.use_tls)
  sensitive = true
}