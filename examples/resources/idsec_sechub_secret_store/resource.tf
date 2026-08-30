# Example: AWS secret store — public access (no connector needed)
# Use this when your AWS Secrets Manager is accessible over the public internet.
# No connection_config block is required; public access is the default.
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

# Example: AWS secret store — private access via connector
# Use this when your AWS Secrets Manager is on a private network and is NOT accessible
# over the public internet. Secrets Hub reaches it through a secure connector.
resource "idsec_sechub_secret_store" "example_aws_connector" {
  name        = "example-aws-secret-store-connector"
  description = "An example Secrets Hub AWS secret store using a connector"
  type        = "AWS_ASM"
  behaviors   = ["SECRETS_TARGET"]

  data = {
    account_alias         = "example-aws-account"
    account_id            = "123456789012"
    region_id             = "us-east-1"
    role_name             = "SecretsHubRole"
    authentication_method = "GLOBAL_ROLE_EXTERNAL_ID"

    # Required when your secret store is on a private network.
    # connection_type must be set to "CONNECTOR".
    # connector_pool_id is the UUID used to connect PAM Self-Hosted and Secrets Hub.
    connection_config = {
      connection_type   = "CONNECTOR"
      connector_pool_id = "00000000-0000-0000-0000-000000000004"
    }
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

    # connection_config is required for Azure. Use connection_type = "CONNECTOR" for private
    # network access (provide connector_pool_id). Use connection_type = "PUBLIC" for public
    # internet access (omit connector_pool_id).
    connection_config = {
      connection_type   = "CONNECTOR"
      connector_pool_id = "00000000-0000-0000-0000-000000000004"
    }
  }
}

# Example: GCP secret store — public access (no connector needed)
# Use this when your GCP Secret Manager is accessible over the public internet.
# No connection_config block is required; public access is the default.
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

# Example: GCP secret store — private access via connector
resource "idsec_sechub_secret_store" "example_gcp_connector" {
  name        = "example-gcp-secret-store-connector"
  description = "An example Secrets Hub GCP secret store using a connector"
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

    # Required when your secret store is on a private network.
    # connection_type must be set to "CONNECTOR".
    # connector_pool_id is the UUID used to connect PAM Self-Hosted and Secrets Hub
    connection_config = {
      connection_type   = "CONNECTOR"
      connector_pool_id = "00000000-0000-0000-0000-000000000004"
    }
  }
}

# Example: HashiCorp Vault secret store — private access via connector (using connector_pool_id)
# connector_pool_id and connector_id are mutually exclusive.
resource "idsec_sechub_secret_store" "example_hashi" {
  name        = "example-hashi-secret-store"
  description = "An example Secrets Hub HashiCorp Vault secret store"
  type        = "HASHICORP_VAULT"
  behaviors   = ["SECRETS_TARGET"]

  data = {
    hashi_vault_url = "https://vault.example.com"
    # Mount path must include a trailing slash, following the HashiCorp Vault convention (e.g. "secret/", "kv/")
    mount_path = "secret/"
    role_name  = "secrets-hub-role"
    # Authentication path must include a trailing slash, following the HashiCorp Vault convention (e.g. "auth/jwt/login/")
    authentication_path = "auth/jwt/login/"

    connection_config = {
      connection_type   = "CONNECTOR"
      connector_pool_id = "00000000-0000-0000-0000-000000000004"
    }
  }
}

# Example: HashiCorp Vault secret store — private access via connector (using connector_id)
# connector_pool_id and connector_id are mutually exclusive.
resource "idsec_sechub_secret_store" "example_hashi_connector_id" {
  name        = "example-hashi-secret-store-cid"
  description = "An example Secrets Hub HashiCorp Vault secret store using connector_id"
  type        = "HASHICORP_VAULT"
  behaviors   = ["SECRETS_TARGET"]

  data = {
    hashi_vault_url     = "https://vault.example.com"
    mount_path          = "secret/"
    role_name           = "secrets-hub-role"
    authentication_path = "auth/jwt/login/"

    connection_config = {
      connection_type = "CONNECTOR"
      connector_id    = "ManagementAgent_90c63827-7315-4284-8559-000000000001"
    }
  }
}

# Example: HashiCorp Vault secret store — public access
# connection_config is required for HashiCorp Vault. Use connection_type = "PUBLIC" when
# your Vault is accessible over the public internet (no connector needed).
resource "idsec_sechub_secret_store" "example_hashi_public" {
  name        = "example-hashi-secret-store-public"
  description = "An example Secrets Hub HashiCorp Vault secret store with public access"
  type        = "HASHICORP_VAULT"
  behaviors   = ["SECRETS_TARGET"]

  data = {
    hashi_vault_url     = "https://vault.example.com"
    mount_path          = "secret/"
    role_name           = "secrets-hub-role"
    authentication_path = "auth/jwt/login/"

    connection_config = {
      connection_type = "PUBLIC"
    }
  }
}
