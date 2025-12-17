# Create a simple AWS account onboarding with SCA service
resource "idsec_cce_aws_account" "simple_example" {
  account_id           = "123456789012"
  account_display_name = "Terraform Onboarded Account"
  deployment_region    = "us-east-1"

  services = [
    {
      service_name = "sca"
      resources = {
        ScaRole = "arn:aws:iam::123456789012:role/scaRole"
      }
    },
  ]
}

