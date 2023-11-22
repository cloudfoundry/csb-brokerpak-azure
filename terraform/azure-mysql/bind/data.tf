locals {
  username = format("%s@%s", random_string.username.result, var.mysql_hostname)
}