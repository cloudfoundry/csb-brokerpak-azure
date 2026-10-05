locals {
  api_server = "https://${var.private_cluster_enabled ? var.private_fqdn : var.fqdn}"
  exec_recipe = {
    version = "v1"
    prerequisites = [
      "Azure CLI authenticated as an identity authorized for this cluster",
      "kubelogin",
      "kubectl",
      "network access to the private API endpoint when private_cluster_enabled is true",
    ]
    commands = [
      "az aks get-credentials --resource-group ${var.resource_group} --name ${var.cluster_name} --overwrite-existing",
      "kubelogin convert-kubeconfig -l azurecli",
      "kubectl cluster-info",
    ]
  }
}

output "cluster_name" {
  value = var.cluster_name
}

output "cluster_id" {
  value = var.cluster_id
}

output "resource_group" {
  value = var.resource_group
}

output "location" {
  value = var.location
}

output "kubernetes_version" {
  value = var.kubernetes_version
}

output "api_server" {
  value = local.api_server
}

output "private_cluster_enabled" {
  value = var.private_cluster_enabled
}

output "oidc_issuer_url" {
  value = var.oidc_issuer_url
}

output "node_resource_group" {
  value = var.node_resource_group
}

output "ttl_expires_at" {
  value = var.ttl_expires_at
}

output "exec_recipe_json" {
  value = jsonencode(local.exec_recipe)
}

output "normalized_binding_json" {
  value = jsonencode({
    version            = "v1"
    provider           = "azure"
    provisioner_family = "azure_aks"
    connection_type    = "management"
    endpoint = {
      api_server = local.api_server
      private    = var.private_cluster_enabled
      region     = var.location
    }
    access = {
      mode        = "azure_identity_exec"
      interactive = true
      expires_at  = var.ttl_expires_at
    }
    grant = {
      kind                 = "caller_azure_rbac"
      least_privilege_unit = "cluster"
    }
    resource = {
      id                  = var.cluster_id
      name                = var.cluster_name
      resource_group      = var.resource_group
      node_resource_group = var.node_resource_group
      kubernetes_version  = var.kubernetes_version
      oidc_issuer_url     = var.oidc_issuer_url
    }
    exec = local.exec_recipe
  })
}
