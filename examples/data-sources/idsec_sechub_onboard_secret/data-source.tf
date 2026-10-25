data "idsec_sechub_onboard_secret" "example_secret" {
  name     = var.secret_name
  store_id = var.store_id
}
