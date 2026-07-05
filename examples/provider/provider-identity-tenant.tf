provider "idsec" {
  auth_method = "identity"
  username    = var.idsec_username
  secret      = var.idsec_secret
  subdomain   = var.idsec_subdomain # Tenant subdomain, e.g. "my-tenant"
}
