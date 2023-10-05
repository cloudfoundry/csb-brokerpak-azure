terraform {
  required_providers {
    csbsqlserver = {
      source  = "cloudfoundry/csbsqlserver"
      version = "1.0.0"
    }
  }
}

provider "csbsqlserver" {
  server   = "localhost"
  port     = 1433
  username = "SA"
  password = "YOUR_ADMIN_PASSWORD_HERE"
  database = "mydb"
  encrypt  = "disable"
}

resource "csbsqlserver_binding" "binding" {
  username = "test_user"
  password = "test_password"
  roles    = ["db_ddladmin", "db_datareader", "db_datawriter", "db_accessadmin"]
}
