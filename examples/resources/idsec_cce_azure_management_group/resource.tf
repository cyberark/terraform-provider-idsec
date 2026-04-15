# Onboard an Azure Management Group
resource "idsec_cce_azure_management_group" "example" {
  entra_id            = var.entra_id
  management_group_id = var.management_group_id
  services            = var.services
  cce_resources       = var.cce_resources
}
