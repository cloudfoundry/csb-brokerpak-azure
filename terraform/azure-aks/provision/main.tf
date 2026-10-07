resource "azurerm_resource_group" "aks" {
  name     = var.resource_group
  location = var.location
  tags     = local.common_tags
}

resource "azurerm_virtual_network" "aks" {
  name                = local.vnet_name
  location            = azurerm_resource_group.aks.location
  resource_group_name = azurerm_resource_group.aks.name
  address_space       = var.vnet_address_space
  tags                = local.common_tags
}

resource "azurerm_subnet" "aks" {
  name                 = local.subnet_name
  resource_group_name  = azurerm_resource_group.aks.name
  virtual_network_name = azurerm_virtual_network.aks.name
  address_prefixes     = var.subnet_address_prefixes
}

resource "azurerm_kubernetes_cluster" "aks" {
  name                = local.cluster_name
  location            = azurerm_resource_group.aks.location
  resource_group_name = azurerm_resource_group.aks.name
  dns_prefix          = substr(replace(local.cluster_name, "_", "-"), 0, 54)
  node_resource_group = local.node_resource_group

  private_cluster_enabled           = var.private_cluster_enabled
  local_account_disabled            = true
  oidc_issuer_enabled               = true
  workload_identity_enabled         = true
  role_based_access_control_enabled = true

  azure_active_directory_role_based_access_control {
    azure_rbac_enabled = true
    tenant_id          = var.azure_tenant_id
  }
  sku_tier = "Free"

  api_server_access_profile {
    authorized_ip_ranges = var.private_cluster_enabled ? null : var.authorized_api_server_cidrs
  }

  default_node_pool {
    name                         = "system"
    type                         = "VirtualMachineScaleSets"
    vm_size                      = var.node_vm_size
    node_count                   = var.node_count
    auto_scaling_enabled         = false
    node_public_ip_enabled       = false
    only_critical_addons_enabled = true
    vnet_subnet_id               = azurerm_subnet.aks.id
  }

  identity {
    type = "SystemAssigned"
  }

  network_profile {
    network_plugin      = "azure"
    network_plugin_mode = "overlay"
    network_data_plane  = "cilium"
    load_balancer_sku   = "standard"
    outbound_type       = "loadBalancer"
    pod_cidr            = var.pod_cidr
    service_cidr        = var.service_cidr
    dns_service_ip      = var.dns_service_ip
  }

  tags = local.common_tags

  lifecycle {
    precondition {
      condition     = var.private_cluster_enabled ? length(var.authorized_api_server_cidrs) == 0 : length(var.authorized_api_server_cidrs) > 0
      error_message = "Private clusters require an empty authorized_api_server_cidrs list; public clusters require at least one explicit management CIDR."
    }

    precondition {
      condition     = var.ttl_hours == 8 && var.node_count == 3
      error_message = "The sandbox-3-node plan requires an 8-hour TTL and exactly three nodes."
    }
  }
}
