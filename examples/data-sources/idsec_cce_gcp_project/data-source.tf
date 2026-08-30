# Example: Read a CCE GCP project
# This example demonstrates how to retrieve project details
# by providing the project onboarding ID

data "idsec_cce_gcp_project" "example" {
  # project onboarding ID (required)
  id = "aaaa1111bbbb2222cccc3333dddd4444"
}
