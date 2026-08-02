# Example: AWS secret store
resource "idsec_sechub_secret_store" "example_aws" {
  name        = "example-aws-secret-store"
  description = "An example Secrets Hub AWS secret store"
  type        = "AWS_ASM"
  behaviors   = ["SECRETS_TARGET"]

  data = {
    account_alias         = "example-aws-account"
    account_id            = "123456789012"
    region_id             = "us-east-1"
    role_name             = "SecretsHubRole"
    authentication_method = "GLOBAL_ROLE_EXTERNAL_ID"
  }
}

# Example: Azure Key Vault secret store (with connector)
resource "idsec_sechub_secret_store" "example_azure" {
  name        = "example-azure-secret-store"
  description = "An example Secrets Hub Azure Key Vault secret store"
  type        = "AZURE_AKV"
  behaviors   = ["SECRETS_TARGET"]

  data = {
    app_client_directory_id = "00000000-0000-0000-0000-000000000001"
    azure_vault_url         = "https://example-vault.vault.azure.net/"
    app_client_id           = "00000000-0000-0000-0000-000000000002"
    subscription_id         = "00000000-0000-0000-0000-000000000003"
    subscription_name       = "example-subscription"
    resource_group_name     = "example-resource-group"
    authentication_method   = "FEDERATED_IDENTITY"

    connection_config = {
      connection_type   = "CONNECTOR"
      connector_pool_id = "00000000-0000-0000-0000-000000000004"
    }
  }
}

# Example: GCP secret store
resource "idsec_sechub_secret_store" "example_gcp" {
  name        = "example-gcp-secret-store"
  description = "An example Secrets Hub GCP secret store"
  type        = "GCP_GSM"
  behaviors   = ["SECRETS_TARGET"]

  data = {
    gcp_project_name   = "example-project"
    gcp_project_number = "123456789012"

    gcp_authentication = {
      gcp_project_number            = "123456789012"
      gcp_workload_identity_pool_id = "example-pool-id"
      gcp_pool_provider_id          = "example-provider-id"
      service_account_email         = "example-sa@example-project.iam.gserviceaccount.com"
      authentication_method         = "GLOBAL_ROLE_EXTERNAL_ID"
    }
  }
}

# Example: HashiCorp Vault secret store (with connector)
resource "idsec_sechub_secret_store" "example_hashi" {
  name        = "example-hashi-secret-store"
  description = "An example Secrets Hub HashiCorp Vault secret store"
  type        = "HASHICORP_VAULT"
  behaviors   = ["SECRETS_TARGET"]

  data = {
    hashi_vault_url     = "https://vault.example.com"
    mount_path          = "secret"
    role_name           = "secrets-hub-role"
    authentication_path = "auth/jwt/login"

    connection_config = {
      connection_type   = "CONNECTOR"
      connector_pool_id = "00000000-0000-0000-0000-000000000004"
    }
  }
}
