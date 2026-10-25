data "idsec_cce_azure_entra" "example" {
  id = "aaaa1111bbbb2222cccc3333dddd4444" # Onboarding ID of the Entra tenant
}

# Map each onboarded service to its deployed version
output "azure_entra_service_versions" {
  value = {
    for service in data.idsec_cce_azure_entra.example.services_data : service.name => service.version
  }
}

# Map each service to the versions it can be upgraded to (populated for top-level workspaces only)
output "azure_entra_available_versions" {
  value = {
    for service in data.idsec_cce_azure_entra.example.services_data : service.name => service.available_versions
  }
}

# Base CCE infrastructure upgrade information (upgrade_available is null for non-top-level workspaces)
output "azure_entra_cce_upgrade" {
  value = {
    cce_version            = data.idsec_cce_azure_entra.example.cce_version
    cce_available_versions = data.idsec_cce_azure_entra.example.cce_available_versions
    upgrade_available      = data.idsec_cce_azure_entra.example.upgrade_available
  }
}
