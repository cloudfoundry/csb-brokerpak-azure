terraform {
  required_providers {
    csbsqlserver = {
      source  = "cloud-service-broker/csbmssqldbrunfailover"
      version = "1.0.0"
    }
  }
}