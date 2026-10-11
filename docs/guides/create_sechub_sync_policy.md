---
page_title: "Create Secrets Hub Sync Policy"
description: |-
  The following workflow describes how to create a Secrets Hub sync policy that syncs secrets from a PAM safe to a target secret store.
---

# Motivation

This workflow describes how to set up a Secrets Hub sync policy using the Idsec Terraform Provider. A sync policy defines which secrets to sync with which PAM safe and where to sync them to (a target secret store such as Azure Key Vault, AWS Secrets Manager, GCP Secret Manager or HashiCorp Vault).

# Workflow

The workflow will:
- Authenticate to Idira with a user who has the required permissions to create terraform resources on your tenant.
- Create a PAM safe (or use an existing one).
- Add Secrets Hub as safe member.
- Create a target secret store (Azure Key Vault in this example).
- Create a sync policy that ties the source PAM store, the safe where your account is and the target store together.

main.tf
```terraform
terraform {
  required_version = ">= 0.13"
  required_providers {
    idsec = {
      source  = "cyberark/idsec"
      version = ">= 0.13"
    }
  }
}

provider "idsec" {
  auth_method = "identity"
  username    = var.idsec_username
  secret      = var.idsec_secret
}

resource "idsec_pcloud_safe" "safe" {
  safe_name                = var.safe_name
  description              = "Safe used as the sync policy source filter"
  number_of_days_retention = var.retention_days
}

# Secrets Hub user must be a safe member with the correct permissions
resource "idsec_pcloud_safe_member" "safe_member" {
  safe_id        = idsec_pcloud_safe.safe.safe_id
  member_name    = var.safe_member_name
  member_type    = "User"
  permission_set = var.safe_member_permission_set
}
        
# same for for AWS, GCP and Hashi
resource "idsec_sechub_secret_store" "target_store" {
  name        = var.target_store_name
  description = var.target_store_description
  type        = "AZURE_AKV"
  behaviors   = ["SECRETS_TARGET"]

  data = {
    app_client_directory_id = var.azure_directory_id
    azure_vault_url         = var.azure_vault_url
    app_client_id           = var.azure_app_client_id
    subscription_id         = var.azure_subscription_id
    subscription_name       = var.azure_subscription_name
    resource_group_name     = var.azure_resource_group_name
    authentication_method   = "FEDERATED_IDENTITY"

    connection_config = {
      connection_type = "PUBLIC"
    }
  }
}

resource "idsec_sechub_sync_policy" "sync_policy" {
  depends_on  = [idsec_pcloud_safe.safe, idsec_sechub_secret_store.target_store]
  name        = var.sync_policy_name
  description = var.sync_policy_description

  source = {
    id = var.pcloud_source_store_id
  }

  target = {
    id = idsec_sechub_secret_store.target_store.id
  }

  filter = {
    type = "PAM_SAFE"
    data = {
      safe_name = idsec_pcloud_safe.safe.safe_name
    }
  }

  transformation = {
    predefined = "password_only_plain_text"
  }
}
```

variables.tf
```terraform
variable "idsec_username" {
  description = "Username for the Idsec provider"
  type        = string
}

variable "idsec_secret" {
  description = "Secret/password for the Idsec provider"
  type        = string
  sensitive   = true
}

variable "safe_name" {
  description = "Name of the PAM safe to create"
  type        = string
}

variable "retention_days" {
  description = "Number of days to retain safe contents"
  type        = number
  default     = 7
}

variable "safe_member_name" {
  description = "Name of the user or group to add as a safe member"
  type        = string
}

variable "safe_member_permission_set" {
  description = "Permission set to grant the safe member"
  type        = string
}

variable "target_store_name" {
  description = "Name of the target secret store"
  type        = string
}

variable "target_store_description" {
  description = "Description of the target secret store"
  type        = string
  default     = "Azure Key Vault target store for Secrets Hub sync"
}

variable "azure_directory_id" {
  description = "Azure Entra ID tenant (directory) ID used by the app registration"
  type        = string
}

variable "azure_vault_url" {
  description = "Azure Key Vault URL (e.g. https://my-vault.vault.azure.net)"
  type        = string
}

variable "azure_app_client_id" {
  description = "Azure app registration (client) ID"
  type        = string
}

variable "azure_subscription_id" {
  description = "Azure subscription ID"
  type        = string
}

variable "azure_subscription_name" {
  description = "Azure subscription display name"
  type        = string
}

variable "azure_resource_group_name" {
  description = "Azure resource group name where the Key Vault resides"
  type        = string
}

variable "pcloud_source_store_id" {
  description = "ID of the existing PAM source secret store to sync from"
  type        = string
}

variable "sync_policy_name" {
  description = "Name of the sync policy"
  type        = string
  default     = "my-sync-policy"
}

variable "sync_policy_description" {
  description = "Description of the sync policy"
  type        = string
  default     = "Syncs secrets from a PAM safe to Azure Key Vault"
}
```
