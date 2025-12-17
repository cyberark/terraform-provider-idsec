resource "idsec_sia_secrets_vm" "example_pcloud_secret" {
  secret_name         = "example_pcloud_secret"
  secret_type         = "PCloudAccount"
  pcloud_account_name = var.account_name
  pcloud_account_safe = var.safe_name
}

resource "idsec_sia_workspaces_target_sets" "example_target_set" {
  name        = var.target_set_address
  type        = "Target"
  secret_type = "PCloudAccount"
  secret_id   = idsec_sia_secrets_vm.example_pcloud_secret.secret_id
}
