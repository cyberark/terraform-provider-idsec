---
page_title: "terraform-provider-idsec - idsec_sechub_secret_store"
subcategory: "Secrets Hub"
description: Manage Secrets Hub secret store resource that represent secret management systems, including their configuration and metadata
---

# idsec_sechub_secret_store (Resource)

Manage Secrets Hub secret store resource that represent secret management systems, including their configuration and metadata
<!-- BEGIN CUSTOM NOTES -->
<!-- END CUSTOM NOTES -->

## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The secret store name.
- `type` (String) The type for the secrets (AWS_ASM,AZURE_AKV,GCP_GSM,HASHICORP_VAULT,HASHICORP_VAULT_ENT,PAM_PCLOUD,PAM_SELF_HOSTED)

### Optional

- `behaviors` (List of String) Whether the secret store is used as a source or a target. There can be only one source secret store per tenant. Valid values: SECRETS_SOURCE, SECRETS_TARGET
- `data` (Attributes) The data of the secret store depends on the secret store type. (see [below for nested schema](#nestedatt--data))
- `description` (String) A description of the secret store.
- `organization_id` (String)
- `state` (String) The secret store state (ENABLED,DISABLED)

### Read-Only

- `created_at` (String) The secret store creation date.
- `created_by` (String) The user who created the secret store.
- `creation_details` (String) Allowed Values: Secrets Hub, Connect Cloud Environment
- `id` (String) The unique identifier of the secret store
- `scan` (Attributes) (see [below for nested schema](#nestedatt--scan))
- `store_status` (Attributes) (see [below for nested schema](#nestedatt--store_status))
- `total_policies_count` (Number) The total amount of policies in the secret store
- `total_secrets_count` (Number) The total amount of secrets in the secret store
- `updated_at` (String) The last date the secret store was updated
- `updated_by` (String) The last user to update the secret store.

<a id="nestedatt--data"></a>
### Nested Schema for `data`

Optional:

- `account_alias` (String) AWS: The alias of your AWS account
- `account_id` (String) AWS: The 12-digit account ID of the AWS account that has the AWS Secrets Manager where you store secrets
- `app_client_directory_id` (String) AZURE: The Azure Active Directory ID of the application that has access to the Azure Key Vault
- `app_client_id` (String) AZURE: The Azure Active Directory application ID of the application that has access to the Azure Key Vault
- `authentication_method` (String) Provider-specific authentication method to use
- `authentication_path` (String) HASHI, HASHI ENT: The authentication path configured in HashiCorp Vault for Secrets Hub to authenticate and access secrets. Example: 'auth/secrets-hub/login' for an authentication path of 'secrets-hub'
- `azure_vault_url` (String) AZURE: The URL of the Azure Key Vault where you store secrets. Example: https://myvault.vault.azure.net
- `connection_config` (Attributes) The network access configuration set for your target (see [below for nested schema](#nestedatt--data--connection_config))
- `connector_id` (String) SELF HOSTED: The connector unique identifier used to connect Secrets Hub and the Cloud Vendor.
- `connector_pool_id` (String) SELF HOSTED: The connector pool unique identifier used to connect PAM Self-Hosted and Secrets Hub.
- `gcp_authentication` (Attributes) GCP: The GCP authentication configuration for the secret store (see [below for nested schema](#nestedatt--data--gcp_authentication))
- `gcp_pool_provider_id` (String) GCP: The GCP pool provider ID created for Secrets Hub to access the GCP Secret Manager
- `gcp_project_name` (String) GCP: The name of the GCP project where the GCP Secret Manager is stored
- `gcp_project_number` (String) GCP: The number of the GCP project where the GCP Secret Manager is stored
- `gcp_workload_identity_pool_id` (String) GCP: The GCP workload identity pool ID created for Secrets Hub to access the GCP Secret Manager
- `hashi_vault_url` (String) HASHI, HASHI ENT: The URL of the HashiCorp Vault where you store secrets. Example: https://myvault.com
- `mount_path` (String) HASHI, HASHI ENT: The mount path of the HashiCorp Vault where secrets are stored. Example: 'secret' for secrets stored in the 'secret' engine
- `namespace` (String) HASHI ENT: The namespace path within HashiCorp Vault used to isolate secrets. Example: root
- `password` (String, Sensitive) SELF HOSTED: The password of the user in PAM 'SecretsHub'
- `region_id` (String) AWS: The region ID for the AWS Secrets Manager
- `resource_group_name` (String) AZURE: The name of the Azure resource group where the Azure Key Vault is stored
- `role_name` (String) COMMON - AWS, HASHI, HASHI ENT: The role used for authentication. For AWS, this is the IAM role ARN. For HashiCorp, this is the role name created in HashiCorp Vault for Secrets Hub to authenticate and access secrets.
- `service_account_email` (String) GCP: The service account email created for Secrets Hub to access the GCP Secret Manager
- `subscription_id` (String) AZURE: The Azure subscription ID where the Azure Key Vault is stored
- `subscription_name` (String) AZURE: The name of the Azure subscription where the Azure Key Vault is stored
- `url` (String) SELF HOSTED: The URL of your PAM Self-Hosted PVWA, or the load balancer for the PVWA
- `username` (String) SELF HOSTED: The user used for Secrets Hub to get secrets from PAM source. Should be 'SecretsHub'. This user should be created by REST API in PAM.

Read-Only:

- `engine_api_version` (String) The API version of the engine in HashiCorp Vault. Valid values: 1, 2
- `engine_type` (String) The type of the engine in HashiCorp Vault. Valid values: KV, PKI, SSH

<a id="nestedatt--data--connection_config"></a>
### Nested Schema for `data.connection_config`

Optional:

- `connection_type` (String) The type of connector (CONNECTOR,PUBLIC)
- `connector_id` (String) AZURE, HASHI, HASHI ENT: The connector unique identifier used to connect Secrets Hub and the Cloud Vendor.
- `connector_pool_id` (String) The connector pool unique identifier used to connect PAM Self-Hosted and Secrets Hub.


<a id="nestedatt--data--gcp_authentication"></a>
### Nested Schema for `data.gcp_authentication`

Optional:

- `authentication_method` (String) GCP: GCP authentication method to use
- `gcp_pool_provider_id` (String) GCP: The GCP pool provider ID created for Secrets Hub to access the GCP Secret Manager
- `gcp_project_number` (String) GCP: The number of the GCP project where the service account is stored. If not provided, defaults to the Secret Store's gcpProjectNumber
- `gcp_workload_identity_pool_id` (String) GCP: The GCP workload identity pool ID created for Secrets Hub to access the GCP Secret Manager
- `service_account_email` (String) GCP: The service account email created for Secrets Hub to access the GCP Secret Manager



<a id="nestedatt--scan"></a>
### Nested Schema for `scan`

Optional:

- `finished_at` (String) The date and time the scan ended. Example: 2023-07-06T15:45:00.103000
- `message` (String) More information on the scan status.
- `status` (String) The status of the scan (IN_PROGRESS,SUCCESS,FAILED)

Read-Only:

- `id` (String) The unique identifier of the scan


<a id="nestedatt--store_status"></a>
### Nested Schema for `store_status`

Read-Only:

- `message` (String) More information on the secret store status.
- `status` (String) The status of the secret store (SUCCESS, FAILED)




## Import

The `idsec_sechub_secret_store` resource can be imported using the following command:

```shell
terraform import idsec_sechub_secret_store.example secret-store-id
```