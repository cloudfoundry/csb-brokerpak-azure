terraform {
  required_providers {
    random = {
      source  = "registry.terraform.io/hashicorp/random"
      version = ">=3.3.1"
    }
    csbsqlserver = {
      source  = "cloud-service-broker/csbsqlserver"
      version = ">=1.0.0"
    }
  }
}