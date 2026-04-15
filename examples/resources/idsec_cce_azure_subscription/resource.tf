# Onboard an Azure Subscription with SCA service
resource "idsec_cce_azure_subscription" "example" {
  entra_id          = var.entra_id
  subscription_id   = var.subscription_id
  entra_tenant_name = var.entra_tenant_name
  subscription_name = var.subscription_name
  services          = var.services
}
