output "username" { value = csbpg_binding_user.new_user.username }
output "password" {
  value     = csbpg_binding_user.new_user.password
  sensitive = true
}
output "uri" {
  value = format("postgresql://%s:%s@%s:%d/%s",
    csbpg_binding_user.new_user.username,
    csbpg_binding_user.new_user.password,
    var.hostname,
    var.port,
  var.db_name)
  sensitive = true
}

output "jdbcUrl" {
  value = format("jdbc:postgresql://%s:%s/%s?user=%s\u0026password=%s\u0026verifyServerCertificate=true\u0026ssl=true\u0026sslmode=require\u0026serverTimezone=GMT",
    var.hostname,
    var.port,
    var.db_name,
    csbpg_binding_user.new_user.username,
    csbpg_binding_user.new_user.password)
  sensitive = true
}