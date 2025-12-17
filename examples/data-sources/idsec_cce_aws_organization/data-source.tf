# Read AWS organization details by onboarding ID
data "idsec_cce_aws_organization" "example" {
  id = var.organization_id # CCE onboarding ID of the organization
}

# Output the full organization object
output "full_organization" {
  description = "Complete organization details"
  value       = data.idsec_cce_aws_organization.example
}

# Output specific organization attributes
output "organization_id" {
  description = "CCE onboarding ID of the organization"
  value       = data.idsec_cce_aws_organization.example.id
}

output "organization_display_name" {
  description = "Display name of the organization"
  value       = data.idsec_cce_aws_organization.example.display_name
}

output "aws_organization_id" {
  description = "AWS Organization ID"
  value       = data.idsec_cce_aws_organization.example.organization_id
}

output "management_account_id" {
  description = "AWS Management Account ID"
  value       = data.idsec_cce_aws_organization.example.management_account_id
}

output "status" {
  description = "Onboarding status of the organization"
  value       = data.idsec_cce_aws_organization.example.status
}

