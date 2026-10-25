# Read AWS organization details by onboarding ID
data "idsec_cce_aws_organization" "example" {
  id = var.organization_id # CCE onboarding ID of the organization
}

# Output the full organization object
output "full_organization" {
  description = "Full organization details"
  value       = data.idsec_cce_aws_organization.example
}

# Output specific organization attributes
output "organization_id" {
  description = "CCE organization onboarding ID"
  value       = data.idsec_cce_aws_organization.example.id
}

output "organization_display_name" {
  description = "Display name of the organization"
  value       = data.idsec_cce_aws_organization.example.display_name
}

output "aws_organization_id" {
  description = "AWS organization ID"
  value       = data.idsec_cce_aws_organization.example.organization_id
}

output "management_account_id" {
  description = "AWS management account ID"
  value       = data.idsec_cce_aws_organization.example.management_account_id
}

output "status" {
  description = "Onboarding status of the organization"
  value       = data.idsec_cce_aws_organization.example.status
}

output "cce_upgrade" {
  description = "Base CCE infrastructure version, the versions it can be upgraded to, and whether any upgrade (CCE or service) is available"
  value = {
    cce_version            = data.idsec_cce_aws_organization.example.cce_version
    cce_available_versions = data.idsec_cce_aws_organization.example.cce_available_versions
    upgrade_available      = data.idsec_cce_aws_organization.example.upgrade_available
  }
}

