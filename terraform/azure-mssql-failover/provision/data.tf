locals {
  instance_types = {
    1  = "GP_Gen5_1"
    2  = "GP_Gen5_2"
    4  = "GP_Gen5_4"
    8  = "GP_Gen5_8"
    16 = "GP_Gen5_16"
    32 = "GP_Gen5_32"
    80 = "GP_Gen5_80"
  }
  sku_name       = length(var.sku_name) == 0 ? local.instance_types[var.cores] : var.sku_name
  resource_group = length(var.resource_group) == 0 ? format("rg-%s", var.instance_name) : var.resource_group

  serverFQDN = format("%s.database.windows.net", azurerm_sql_failover_group.failover_group.name)

  default_pair = {
    // https://docs.microsoft.com/en-us/azure/best-practices-availability-paired-regions
    "eastasia"           = "southeastasia"
    "southeastasia"      = "eastasia"
    "centralus"          = "eastus2"
    "eastus"             = "westus"
    "eastus2"            = "centralus"
    "westus"             = "eastus"
    "northcentralus"     = "southcentralus"
    "southcentralus"     = "northcentralus"
    "northeurope"        = "westeurope"
    "westeurope"         = "northeurope"
    "japanwest"          = "japaneast"
    "japaneast"          = "japanwest"
    "brazilsouth"        = "southcentralus"
    "australiaeast"      = "australiasoutheast"
    "australiasoutheast" = "australiaeast"
    "australiacentral"   = "australiacentral2"
    "australiacentral2"  = "australiacentral"
    "southindia"         = "centralindia"
    "centralindia"       = "southindia"
    "westindia"          = "southindia"
    "canadacentral"      = "canadaeast"
    "canadaeast"         = "canadacentral"
    "uksouth"            = "ukwest"
    "ukwest"             = "uksouth"
    "westcentralus"      = "westus2"
    "westus2"            = "westcentralus"
    "koreacentral"       = "koreasouth"
    "koreasouth"         = "koreacentral"
    "francecentral"      = "francesouth"
    "francesouth"        = "francecentral"
    "uaenorth"           = "uaecentral"
    "uaecentral"         = "uaenorth"
    "southafricanorth"   = "southafricawest"
    "southafricawest"    = "southafricanorth"
    "germanycentral"     = "germanynortheast"
    "germanynortheast"   = "germanycentral"
  }
}
