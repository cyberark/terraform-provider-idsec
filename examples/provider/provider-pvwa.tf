provider "idsec" {
  auth_method       = "pvwa"
  pvwa_url          = var.pvwa_url
  pvwa_login_method = "cyberark" # Options: cyberark, ldap, windows
  username          = var.pvwa_username
  secret            = var.pvwa_password
}
