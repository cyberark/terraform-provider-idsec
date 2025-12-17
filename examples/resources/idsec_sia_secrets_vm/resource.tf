resource "idsec_sia_secrets_vm" "example_provisioner_secret" {
  secret_name          = "example_provisioner_secret"
  secret_type          = "ProvisionerUser"
  provisioner_username = var.provisioner_username
  provisioner_password = var.provisioner_password
}

resource "idsec_sia_secrets_vm" "example_pcloud_secret" {
  secret_name         = "example_pcloud_secret"
  secret_type         = "PCloudAccount"
  pcloud_account_name = var.account_name
  pcloud_account_safe = var.safe_name
}
