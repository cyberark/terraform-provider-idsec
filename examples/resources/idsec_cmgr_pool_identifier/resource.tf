resource "idsec_cmgr_pool_identifier" "example_identifier" {
  type    = "GENERAL_FQDN"
  value   = var.pool_identifier_fqdn
  pool_id = var.pool_id
}
