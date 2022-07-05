provider "csbmssqldbrunfailover" {
  azure_tenant_id       = "some-tenant-id"
  azure_client_id       = "some-client-id"
  azure_client_secret   = "client-secret"
  azure_subscription_id = "subscription-id"
}

resource "csbmssqldbrunfailover_failover" "failover" {
  resource_group                = "resource-group"
  partner_server_resource_group = "partner-server-resource-group"
  server_name                   = "server-name"
  partner_server_name           = "partner-server-name"
  failover_group                = "failover-group"
}