output "cluster_name" {
  value = azurerm_kubernetes_cluster.aks.name
}

output "cluster_id" {
  value = azurerm_kubernetes_cluster.aks.id
}

output "resource_group" {
  value = azurerm_resource_group.aks.name
}

output "location" {
  value = azurerm_resource_group.aks.location
}

output "kubernetes_version" {
  value = azurerm_kubernetes_cluster.aks.kubernetes_version
}

output "fqdn" {
  value = azurerm_kubernetes_cluster.aks.fqdn
}

output "private_fqdn" {
  value = azurerm_kubernetes_cluster.aks.private_fqdn
}

output "private_cluster_enabled" {
  value = var.private_cluster_enabled
}

output "oidc_issuer_url" {
  value = azurerm_kubernetes_cluster.aks.oidc_issuer_url
}

output "node_resource_group" {
  value = azurerm_kubernetes_cluster.aks.node_resource_group
}

output "ttl_expires_at" {
  value = local.ttl_expires_at
}

output "normalized_instance_json" {
  value = jsonencode({
    version            = "v1"
    provider           = "azure"
    provisioner_family = "azure_aks"
    resource = {
      id                  = azurerm_kubernetes_cluster.aks.id
      name                = azurerm_kubernetes_cluster.aks.name
      resource_group      = azurerm_resource_group.aks.name
      node_resource_group = azurerm_kubernetes_cluster.aks.node_resource_group
      region              = azurerm_resource_group.aks.location
      type                = "managed_kubernetes_cluster"
    }
    endpoint = {
      api_server = local.api_server
      private    = var.private_cluster_enabled
    }
    kubernetes = {
      version                   = azurerm_kubernetes_cluster.aks.kubernetes_version
      node_count                = var.node_count
      node_vm_size              = var.node_vm_size
      autoscaling_enabled       = false
      node_public_ip_enabled    = false
      network_plugin            = "azure"
      network_plugin_mode       = "overlay"
      network_data_plane        = "cilium"
      load_balancer_sku         = "standard"
      oidc_issuer_url           = azurerm_kubernetes_cluster.aks.oidc_issuer_url
      workload_identity_enabled = true
    }
    lifecycle = {
      ttl_hours  = var.ttl_hours
      expires_at = local.ttl_expires_at
    }
  })
}
