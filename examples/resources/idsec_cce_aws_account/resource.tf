# Create a simple AWS account onboarding with the DPA service
resource "idsec_cce_aws_account" "simple_example" {
  account_id           = "123456789012"
  account_display_name = "Terraform onboarded account"
  deployment_region    = "us-east-1"

  services = [
    {
      service_name = "dpa"
      version      = "0.0.2"
      resources = {
        DpaRoleArn = "arn:aws:iam::123456789012:role/CyberArkDynamicPrivilegedAccess"
      }
    },
  ]
}
