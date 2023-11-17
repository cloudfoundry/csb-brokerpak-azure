moved {
  from = module.instance.mysql_grant.newuser
  to   = mysql_grant.newuser
}
moved {
  from = module.instance.random_string.username
  to   = random_string.username
}
moved {
  from = module.instance.random_password.password
  to   = random_password.password
}
moved {
  from = module.instance.mysql_user.newuser
  to   = mysql_user.newuser
}
