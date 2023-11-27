terraform {
  required_providers {
    random = {
      source  = "hashicorp/random"
      version = ">=3.3.1"
    }

    azurerm = {
      source  = "hashicorp/azurerm"
      version = ">=3.81.0"
    }
  }
}

