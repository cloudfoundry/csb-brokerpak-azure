locals {
  tags           = merge(var.labels, { "heritage" : "cloud-service-broker" })
  resource_group = length(var.resource_group) == 0 ? format("rg-%s", var.instance_name) : var.resource_group
}