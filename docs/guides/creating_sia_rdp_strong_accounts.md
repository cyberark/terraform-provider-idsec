---
page_title: "Creating SIA-RDP Strong Accounts"
description: |-
  Creates and configures VM secrets (strong accounts) for secure RDP connections to Windows servers.
---

# Creating SIA-RDP Strong Accounts

## Motivation

When an end user logs in to a Windows machine (a protected platform resource) using Remote Desktop Access, SIA creates an ephemeral user either on the local machine (in C:\ Users) or on the domain controller depending on how access has been configured. This ephemeral user gives the end user the necessary permissions to perform their work, and the user profile is deleted from the machine when the end user logs out. To provision this ephemeral user, SIA needs to access a strong account (VM secret), which can in turn create the ephemeral users. For more, see https://docs.cyberark.com/ispss-access/latest/en/content/introduction/dpa_strong-account.htm


The following workflow describes how to create and configure VM secrets for SIA-RDP, including various account types and domain configurations.

## Understanding Strong Accounts

### Secret Types

#### ProvisionerUser
Credentials stored directly in CyberArk's secrets service.
- **Best For**: Environments with no pCloud.
- **Requires**: Username and password.

#### PCloudAccount
Reference to credentials stored in Privilege Cloud vault.
- **Best For**: Production environments requiring compliance, password rotation, and centralized management
- **Requires**: Account name and safe name in Privilege Cloud

### Notes (How the provider models secrets)

- **`secret_details` defaults**: `secret_details` is a JSON string. The provider merges your JSON with defaults that include `account_domain` (defaults to `"local"`) and other fields. It’s still a good idea to set `account_domain` explicitly so intent is obvious in code review.
- **`secret_name` for `PCloudAccount`**: the provider schema notes that for `PCloudAccount` the name is auto-generated from the Privilege Cloud account + safe. Keep `secret_name` descriptive in Terraform, but expect the API/provider to derive the actual display name.

### Account Domain Options

| Domain Type | Example | Use Case |
|-------------|---------|----------|
| `local` | Local administrator | Standalone servers, workgroup machines |
| Domain name | `MYDOMAIN`, `corp.local` | Domain-joined machines using domain credentials |

---

## Workflow

The workflow demonstrates creating strong accounts for different scenarios:
- Local administrator accounts
- Domain administrator accounts
- Accounts with different credential formats

## Basic Strong Account (Local Admin)

main.tf
```terraform
terraform {
  required_version = ">= 0.13"
  required_providers {
    idsec = {
      source  = "cyberark/idsec"
      version = ">= 0.1"
    }
  }
}

provider "idsec" {
  auth_method = "identity"
  username    = var.idsec_username
  secret      = var.idsec_secret
}

# Local Administrator Strong Account
resource "idsec_sia_secrets_vm" "local_admin" {
  secret_name          = var.secret_name
  secret_type          = "ProvisionerUser"
  provisioner_username = var.provisioner_username
  provisioner_password = var.provisioner_password

  # Optional: Account metadata
  secret_details = jsonencode({
    account_domain = "local"
    description    = var.secret_description
  })
}
```

variables.tf
```terraform
variable "idsec_username" {
  description = "Username for the Idsec provider (must have DpaAdmin role)"
  type        = string
}

variable "idsec_secret" {
  description = "Secret/password for the Idsec provider"
  type        = string
  sensitive   = true
}

variable "secret_name" {
  description = "Name of the strong account"
  type        = string
  default     = "LocalAdmin-RDP"
}

variable "provisioner_username" {
  description = "Username for the RDP account"
  type        = string
  default     = "Administrator"
}

variable "provisioner_password" {
  description = "Password for the RDP account"
  type        = string
  sensitive   = true
}

variable "secret_description" {
  description = "Description/purpose of the strong account"
  type        = string
  default     = "Local administrator account for RDP access"
}
```

---

## Domain Administrator Strong Account

main.tf
```terraform
terraform {
  required_version = ">= 0.13"
  required_providers {
    idsec = {
      source  = "cyberark/idsec"
      version = ">= 0.1"
    }
  }
}

provider "idsec" {
  auth_method = "identity"
  username    = var.idsec_username
  secret      = var.idsec_secret
}

# Domain Administrator Strong Account
resource "idsec_sia_secrets_vm" "domain_admin" {
  secret_name          = var.secret_name
  secret_type          = "ProvisionerUser"
  provisioner_username = var.provisioner_username
  provisioner_password = var.provisioner_password

  secret_details = jsonencode({
    account_domain = var.domain_name
    description    = var.secret_description
  })
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

variable "secret_name" {
  description = "Name of the strong account"
  type        = string
  default     = "DomainAdmin-RDP"
}

variable "domain_name" {
  description = "Active Directory domain name (e.g., MYDOMAIN, corp.local)"
  type        = string
}

variable "provisioner_username" {
  description = "Username for the domain RDP account"
  type        = string
  default     = "Administrator"
}

variable "provisioner_password" {
  description = "Password for the domain RDP account"
  type        = string
  sensitive   = true
}

variable "secret_description" {
  description = "Description/purpose of the strong account"
  type        = string
  default     = "Domain administrator account for SIA-RDP provisioning"
}
```

---

## Privilege Cloud Account Reference

For production environments, reference credentials stored in Privilege Cloud:

main.tf
```terraform
terraform {
  required_version = ">= 0.13"
  required_providers {
    idsec = {
      source  = "cyberark/idsec"
      version = ">= 0.1"
    }
  }
}

provider "idsec" {
  auth_method = "identity"
  username    = var.idsec_username
  secret      = var.idsec_secret
}

# Privilege Cloud Account Reference
resource "idsec_sia_secrets_vm" "pcloud_account" {
  secret_name         = var.secret_name
  secret_type         = "PCloudAccount"
  pcloud_account_name = var.pcloud_account_name
  pcloud_account_safe = var.pcloud_safe_name

  secret_details = jsonencode({
    account_domain = var.domain_name
    description    = var.secret_description
  })
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

variable "secret_name" {
  description = "Name of the strong account"
  type        = string
  default     = "PCloudAdmin-RDP"
}

variable "pcloud_account_name" {
  description = "Account name in Privilege Cloud vault"
  type        = string
}

variable "pcloud_safe_name" {
  description = "Safe name in Privilege Cloud containing the account"
  type        = string
}

variable "domain_name" {
  description = "Domain associated with the account"
  type        = string
}

variable "secret_description" {
  description = "Description/purpose of the strong account"
  type        = string
  default     = "Privilege Cloud account for production RDP access"
}
```

---

# Username Format Examples

The `provisioner_username` field supports various formats:

| Format | Example | Description |
|--------|---------|-------------|
| Simple | `Administrator` | Account name only (domain from secret_details) |
| UPN | `admin@corp.example.com` | User Principal Name format |
| Down-level | `MYDOMAIN\admin` | Domain\Username format |
| Email-like | `svc_rdp@example.local` | Service account format |

---

# Managing Account Lifecycle

## Rotating Credentials

Update the password by changing the `provisioner_password`:

```terraform
resource "idsec_sia_secrets_vm" "admin_account" {
  secret_name          = "admin-account"
  secret_type          = "ProvisionerUser"
  provisioner_username = "admin@corp.local"
  provisioner_password = var.new_password  # Updated password

  secret_details = jsonencode({
    account_domain = "corp.local"
    description    = "Credentials rotated on 2024-01-15"
  })
}
```

---

# Best Practices

1. **Use descriptive names**: Include purpose or environment in `secret_name` (e.g., `Prod-WebServers-RDP`)
2. **Use PCloudAccount for production**: Leverage Privilege Cloud's password rotation and audit features
3. **Never hardcode passwords**: Always use variables with `sensitive = true`
4. **Add meaningful descriptions**: Document the account's purpose in `secret_details`

---

# Related Resources

- [Creating RDP Target Sets](creating_sia_rdp_target_sets.md) - Associate strong accounts with Windows targets
- [Managing RDP Target Sets and Secrets](managing_rdp_target_sets_and_secrets.md) - Comprehensive management guide
- [Provisioning RDP Access](provisioning_rdp_access.md) - End-to-end RDP access setup

---

# Additional Information

For more details on the resources used in this workflow:

- **Resource**: [idsec_sia_secrets_vm](../resources/sia/sia_secrets_vm.md)
- **Data Source**: [idsec_sia_secrets_vm](../data-sources/sia/sia_secrets_vm.md)

