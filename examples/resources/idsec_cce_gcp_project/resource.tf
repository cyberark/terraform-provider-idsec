# Onboard a GCP project with the SCA service
resource "idsec_cce_gcp_project" "example" {
  project_id      = "my-gcp-project-12345"
  organization_id = "123456789012"
  project_number  = "987654321098"

  services = [
    {
      service_name = "sca"
      resources = {
        workloadIdentityPoolId     = "sca-pool-id"
        workloadIdentityProviderId = "sca-provider-id"
        targetServiceAccountEmail  = "sca-service-account@my-gcp-project-12345.iam.gserviceaccount.com"
        projectNumber              = "987654321098"
        identityTrustedUsername    = "SCA_ISOLATED_SYSTEM_USER_FOR_GCP_123456789ABC@CYBERARK.CLOUD.408498"
      }
    },
  ]
}
