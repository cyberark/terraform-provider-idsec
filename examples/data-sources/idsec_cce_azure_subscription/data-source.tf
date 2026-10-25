data "idsec_cce_azure_subscription" "example" {
  id = "bbbb2222cccc3333dddd4444eeee5555" # Onboarding ID of the subscription
}

# Map each onboarded service to its deployed version
output "azure_subscription_service_versions" {
  value = {
    for service in data.idsec_cce_azure_subscription.example.services_data : service.name => service.version
  }
}

# Map each service to the versions it can be upgraded to (populated for top-level workspaces only)
output "azure_subscription_available_versions" {
  value = {
    for service in data.idsec_cce_azure_subscription.example.services_data : service.name => service.available_versions
  }
}

# Base CCE infrastructure upgrade information (upgrade_available is null for non-top-level workspaces)
output "azure_subscription_cce_upgrade" {
  value = {
    cce_version            = data.idsec_cce_azure_subscription.example.cce_version
    cce_available_versions = data.idsec_cce_azure_subscription.example.cce_available_versions
    upgrade_available      = data.idsec_cce_azure_subscription.example.upgrade_available
  }
}
