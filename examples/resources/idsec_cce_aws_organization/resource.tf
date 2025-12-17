# Create an AWS organization onboarding with multiple services
resource "idsec_cce_aws_organization" "organization_example" {
  # AWS Organization identifiers
  organization_root_id  = var.organization_root_id
  management_account_id = var.management_account_id
  organization_id       = var.organization_id
  region                = var.region

  # Display name for the organization in CCE
  organization_display_name = var.organization_display_name

  # IAM role configuration for scanning
  scan_organization_role_arn     = var.scan_organization_role_arn
  cross_account_role_external_id = var.cross_account_role_external_id

  # Multiple services configuration
  services = var.services
}

# Output the organization ID for reference
output "organization_id" {
  description = "CCE onboarding ID of the organization"
  value       = idsec_cce_aws_organization.organization_example.id
}

output "organization_display_name" {
  description = "Display name of the organization"
  value       = idsec_cce_aws_organization.organization_example.organization_display_name
}


# Add an AWS account to the organization
resource "idsec_cce_aws_organization_account" "organization_account_example" {
  # Reference to the parent organization (CCE onboarding ID)
  parent_organization_id = idsec_cce_aws_organization.organization_example.id

  # AWS Account ID to add to the organization
  account_id = var.account_id

  # Multiple services configuration
  services = var.services
}