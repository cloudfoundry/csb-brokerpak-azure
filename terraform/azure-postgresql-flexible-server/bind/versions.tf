terraform {
  required_providers {
    postgresql = {
      source  = "registry.terraform.io/cyrilgdn/postgresql"
      version = ">=1.16.0"
    }
    random = {
      source  = "registry.terraform.io/hashicorp/random"
      version = ">=3.3.1"
    }
  }
}