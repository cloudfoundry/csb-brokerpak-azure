locals {
  resource_group = length(var.resource_group) == 0 ? format("rg-%s", var.instance_name) : var.resource_group
  admin_password = length(var.admin_password) == 0 ? random_password.password.result : var.admin_password
  admin_username = length(var.admin_username) == 0 ? random_string.username.result : var.admin_username
}