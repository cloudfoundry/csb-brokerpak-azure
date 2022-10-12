variable "db_name" { type = string }
variable "server" { type = string }
variable "server_credentials" { type = map(any) }
variable "labels" { type = map(any) }
variable "sku_name" { type = string }
variable "cores" { type = number }
variable "max_storage_gb" { type = number }
variable "short_term_retention_days" { type = number }
