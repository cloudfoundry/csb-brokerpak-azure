terraform {
  required_providers {
    mysql = {
      source  = "registry.terraform.io/hashicorp/mysql"
      version = ">=1.9.0"
    }
    random = {
      source  = "registry.terraform.io/hashicorp/random"
      version = ">=3.3.1"
    }
  }
}