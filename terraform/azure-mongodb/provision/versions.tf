terraform {
  required_providers {
    azurerm = {
      source  = "registry.terraform.io/hashicorp/azurerm"
      version = "~> 3"
    }
    random = {
      source  = "registry.terraform.io/hashicorp/random"
      version = "~> 3"
    }
  }
}