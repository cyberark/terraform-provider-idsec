data "idsec_cce_azure_management_group" "example" {
  id = "cccc3333dddd4444eeee5555ffff6666" # Onboarding ID of the management group
}

# Map each onboarded service to its deployed version
output "azure_management_group_service_versions" {
  value = {
    for service in data.idsec_cce_azure_management_group.example.services_data : service.name => service.version
  }
}

# Map each service to the versions it can be upgraded to (populated for top-level workspaces only)
output "azure_management_group_available_versions" {
  value = {
    for service in data.idsec_cce_azure_management_group.example.services_data : service.name => service.available_versions
  }
}

# Base CCE infrastructure upgrade information (upgrade_available is null for non-top-level workspaces)
output "azure_management_group_cce_upgrade" {
  value = {
    cce_version            = data.idsec_cce_azure_management_group.example.cce_version
    cce_available_versions = data.idsec_cce_azure_management_group.example.cce_available_versions
    upgrade_available      = data.idsec_cce_azure_management_group.example.upgrade_available
  }
}
