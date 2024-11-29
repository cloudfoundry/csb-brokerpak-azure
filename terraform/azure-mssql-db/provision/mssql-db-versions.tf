terraform {
  required_providers {
    random = {
      source  = "registry.terraform.io/hashicorp/random"
      version = "~> 3"
    }

    azurerm = {
      source  = "registry.terraform.io/hashicorp/azurerm"
      version = "~> 3"
    }
  }
}

