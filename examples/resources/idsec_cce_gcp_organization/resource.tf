# Onboard a GCP organization with the SCA service
resource "idsec_cce_gcp_organization" "example" {
  organization_id       = "123456789012"
  deployment_project_id = "my-hub-project-12345"
  project_number        = "987654321098"

  cce_resources = {
    workload_identity_pool_id     = "org-pool-id"
    workload_identity_provider_id = "org-provider-id"
    target_service_account_email  = "org-service-account@my-hub-project-12345.iam.gserviceaccount.com"
  }

  services = [
    {
      service_name = "sca"
      resources = {
        workloadIdentityPoolId     = "sca-pool-id"
        workloadIdentityProviderId = "sca-provider-id"
        targetServiceAccountEmail  = "sca-service-account@my-hub-project-12345.iam.gserviceaccount.com"
        projectNumber              = "987654321098"
        identityTrustedUsername    = "SCA_ISOLATED_SYSTEM_USER_FOR_GCP_123456789ABC@CYBERARK.CLOUD.408498"
      }
    },
  ]
}
