
resource "random_string" "username" {
  length   = 16
  special  = false
  numeric  = false
}

resource "random_password" "password" {
  length           = 64
  override_special = "~_-."
  min_upper        = 2
  min_lower        = 2
  min_special      = 2
}