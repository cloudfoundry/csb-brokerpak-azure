locals {
  resource_group           = length(var.resource_group) == 0 ? format("rg-%s", var.instance_name) : var.resource_group
  private_endpoint_enabled = var.private_endpoint_subnet_id == null ? false : length(var.private_endpoint_subnet_id) > 0 ? true : false
}