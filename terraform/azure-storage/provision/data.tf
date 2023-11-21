locals {
  resource_group = length(var.resource_group) == 0 ? format("rg-%s", random_string.account_name.result) : var.resource_group
}