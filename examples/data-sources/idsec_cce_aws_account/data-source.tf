data "idsec_cce_aws_account" "example" {
  id = "aaaa1111bbbb2222cccc3333dddd4444" # CCE account onboarding ID
}

# Output the full data object
output "full_account" {
  value = data.idsec_cce_aws_account.example
}

# Map each onboarded service to its deployed version
output "aws_account_service_versions" {
  value = {
    for service in data.idsec_cce_aws_account.example.services_data : service.name => service.version
  }
}

# Map each service to the versions it can be upgraded to (populated for top-level workspaces only)
output "aws_account_available_versions" {
  value = {
    for service in data.idsec_cce_aws_account.example.services_data : service.name => service.available_versions
  }
}

# Base CCE infrastructure upgrade information (upgrade_available is null for non-top-level workspaces)
output "aws_account_cce_upgrade" {
  value = {
    cce_version            = data.idsec_cce_aws_account.example.cce_version
    cce_available_versions = data.idsec_cce_aws_account.example.cce_available_versions
    upgrade_available      = data.idsec_cce_aws_account.example.upgrade_available
  }
}
