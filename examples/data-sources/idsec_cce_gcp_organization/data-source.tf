# Example: Read a CCE GCP organization
# This example demonstrates how to retrieve organization details
# by providing the organization onboarding ID

data "idsec_cce_gcp_organization" "example" {
  # organization onboarding ID (required)
  id = "aaaa1111bbbb2222cccc3333dddd4444"
}
