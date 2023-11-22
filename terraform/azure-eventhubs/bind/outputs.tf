
output "event_hub_connection_string" {
  value     = "${data.azurerm_eventhub_namespace.ns.default_primary_connection_string};EntityPath=${var.eventhub_name}"
  sensitive = true
}

output "event_hub_name" {
  value = var.eventhub_name
}

output "namespace_connection_string" {
  value     = data.azurerm_eventhub_namespace.ns.default_primary_connection_string
  sensitive = true
}

output "namespace_name" {
  value = data.azurerm_eventhub_namespace.ns.name
}

output "shared_access_key_name" {
  value     = "RootManageSharedAccessKey"
  sensitive = true
}

output "shared_access_key_value" {
  value     = data.azurerm_eventhub_namespace.ns.default_primary_key
  sensitive = true
}
