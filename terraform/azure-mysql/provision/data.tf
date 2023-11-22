locals {
  instance_types = {
    1  = "GP_Gen5_1"
    2  = "GP_Gen5_2"
    4  = "GP_Gen5_4"
    8  = "GP_Gen5_8"
    16 = "GP_Gen5_16"
    32 = "GP_Gen5_32"
    64 = "GP_Gen5_64"
  }
  sku_name                 = length(var.sku_name) == 0 ? local.instance_types[var.cores] : var.sku_name
  resource_group           = length(var.resource_group) == 0 ? format("rg-%s", var.instance_name) : var.resource_group
  tls_version              = var.use_tls == true ? var.tls_min_version : "TLSEnforcementDisabled"
  private_endpoint_enabled = var.private_endpoint_subnet_id == null ? false : length(var.private_endpoint_subnet_id) > 0 ? true : false
}