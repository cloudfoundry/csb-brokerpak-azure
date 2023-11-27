terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = ">=3.81.0"
    }
    random = {
      source  = "hashicorp/random"
      version = ">=3.3.1"
    }
  }
}