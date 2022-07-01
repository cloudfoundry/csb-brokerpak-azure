provider "csbmssqldbrunfailover" {
  azure_tenant_id   = "some-tenant-id"
  azure_client_id     = "some-client-id"
  azure_client_secret = "client-secret"
  azure_subscription_id = "subscription-id"
}

resource "csbmssqldbrunfailover" "run_failover" {
  resource_group = "resource-group"
  server_name = "server-name"
  failover_group    = "failover-group"
}
