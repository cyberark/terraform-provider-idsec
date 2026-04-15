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

# Ephemeral Domain User Provisioner
# Creates temporary domain users in Active Directory for RDP sessions
resource "idsec_sia_secrets_vm" "example_ephemeral_domain_secret" {
  secret_name          = "example_ephemeral_domain_provisioner"
  secret_type          = "ProvisionerUser"
  provisioner_username = var.provisioner_username
  provisioner_password = var.provisioner_password

  # Domain settings (required for ephemeral domain users)
  account_domain = "MYDOMAIN"

  # Enable ephemeral domain user creation - REQUIRED
  enable_ephemeral_domain_user_creation = true

  # Domain controller configuration
  domain_controller_name                          = "dc01.mydomain.local"
  domain_controller_netbios                       = "MYDOMAIN"
  domain_controller_use_ldaps                     = true
  domain_controller_enable_certificate_validation = true
  domain_controller_ldaps_certificate             = file("path/to/ldaps-certificate.pem")

  # OU where ephemeral users will be created
  ephemeral_domain_user_location = "OU=myou,DC=mydomain,DC=local"

  # WinRM settings
  use_winrm_for_https                 = true
  winrm_enable_certificate_validation = true
  winrm_certificate                   = file("path/to/winrm-certificate.pem")
}
