# Onboard an Azure Entra tenant
resource "idsec_cce_azure_entra" "example" {
  entra_id      = var.entra_id
  services      = var.services
  cce_resources = var.cce_resources
}
