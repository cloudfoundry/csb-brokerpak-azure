terraform {
  required_providers {
    csbpg = {
      source  = "cloudfoundry.org/cloud-service-broker/csbpg"
      version = ">= 1.0.1"
    }
    random = {
      source  = "registry.terraform.io/hashicorp/random"
      version = "~> 3"
    }
  }
}