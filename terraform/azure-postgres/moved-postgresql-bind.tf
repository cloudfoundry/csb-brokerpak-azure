moved {
  from = module.instance.random_password.password
  to   = random_password.password
}
moved {
  from = module.instance.postgresql_role.new_user
  to   = postgresql_role.new_user
}
moved {
  from = module.instance.postgresql_grant.all_access
  to   = postgresql_grant.all_access
}
moved {
  from = module.instance.postgresql_grant.table_access
  to   = postgresql_grant.table_access
}
moved {
  from = module.instance.random_string.username
  to   = random_string.username
}
