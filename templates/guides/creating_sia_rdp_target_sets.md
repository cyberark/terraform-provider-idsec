---
page_title: "Creating SIA-RDP Target Sets"
description: |-
  Creates and configures target sets for secure RDP access to Windows servers using different targeting strategies.
---

# Creating SIA-RDP Target Sets

## Motivation

Target sets define Windows targets (domains, specific targets, or suffixes) for secure RDP access through CyberArk's Secure Infrastructure Access. Each target set is associated with a strong account.

The following workflow describes how to create target sets for various scenarios, from enterprise-wide domain access to individual server management.

## Understanding Target Set Types

| Type | Name Format | Use Case |
|------|-------------|----------|
| **Domain** | `corp.local`, `MYDOMAIN` | All machines in an Active Directory domain |
| **Suffix** | `*.web.example.com` | Groups of servers matching a wildcard pattern |
| **Target** | `server01.example.com` or `192.168.1.100` | Specific server by FQDN or IP address |

### Notes (How the provider models target sets)

- **Target sets must reference a secret**: `idsec_sia_workspaces_target_set` requires a `secret_id` (from `idsec_sia_secrets_vm.secret_id`). If you delete the secret first, the target set deletion may fail. In Terraform, this usually “just works” due to the implicit reference, but if you ever decouple them, keep the deletion order in mind.
- **`type` drives `name` format**:
  - **Domain**: AD domain name (e.g., `corp.local` or `MYDOMAIN`)
  - **Suffix**: wildcard DNS pattern (e.g., `*.web.corp.local`)
  - **Target**: single FQDN or IP

---

## Workflow

The workflow demonstrates creating target sets for different scenarios. All target sets require an existing strong account (VM secret).

### Prerequisites

Before creating target sets, you need a strong account. See [Creating RDP Strong Accounts](creating_sia_rdp_strong_accounts.md) for details.

---

### Domain Target Set

Enables RDP access to all machines in an Active Directory domain.

main.tf
```terraform
--8<-- "terraform-block.md"

provider "idsec" {
  auth_method = "identity"
  username    = var.idsec_username
  secret      = var.idsec_secret
}

# Strong Account for RDP connections
resource "idsec_sia_secrets_vm" "domain_admin" {
  secret_name          = var.secret_name
  secret_type          = "ProvisionerUser"
  provisioner_username = var.provisioner_username
  provisioner_password = var.provisioner_password

  secret_details = jsonencode({
    account_domain = var.domain_name
    description    = "Domain admin for RDP provisioning"
  })
}

# Domain Target Set
resource "idsec_sia_workspaces_target_set" "domain" {
  name        = var.domain_name
  type        = "Domain"
  description = var.target_set_description

  # Associate with the strong account
  secret_type =  idsec_sia_secrets_vm.domain_admin.secret_type
  secret_id   = idsec_sia_secrets_vm.domain_admin.secret_id

  # Optional: Ephemeral user naming format
  provision_format = var.provision_format

  # Optional: Validate server certificates
  enable_certificate_validation = var.enable_certificate_validation
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

variable "domain_name" {
  description = "Active Directory domain name (e.g., corp.local, MYDOMAIN)"
  type        = string
}

variable "secret_name" {
  description = "Name of the strong account"
  type        = string
}

variable "provisioner_username" {
  description = "Username for the domain RDP account"
  type        = string
}

variable "provisioner_password" {
  description = "Password for the domain RDP account"
  type        = string
  sensitive   = true
}

variable "target_set_description" {
  description = "Description of the target set"
  type        = string
  default     = "Active Directory domain for SIA-RDP access"
}

variable "provision_format" {
  description = "Username format for ephemeral provisioning (e.g., 'eph_{guid}')"
  type        = string
  default     = ""
}

variable "enable_certificate_validation" {
  description = "Enable certificate validation for connections"
  type        = bool
  default     = false
}
```

---

### Suffix Target Set

Enables RDP access to servers matching a wildcard pattern.

main.tf
```terraform
--8<-- "terraform-block.md"

provider "idsec" {
  auth_method = "identity"
  username    = var.idsec_username
  secret      = var.idsec_secret
}

# Strong Account for RDP connections
resource "idsec_sia_secrets_vm" "provisioner" {
  secret_name          = var.secret_name
  secret_type          = "ProvisionerUser"
  provisioner_username = var.provisioner_username
  provisioner_password = var.provisioner_password

  secret_details = jsonencode({
    account_domain = var.account_domain
    description    = "Service account for ${var.suffix_pattern} servers"
  })
}

# Suffix Target Set (Wildcard Pattern)
resource "idsec_sia_workspaces_target_set" "suffix" {
  name        = var.suffix_pattern
  type        = "Suffix"
  description = var.target_set_description

  secret_type = "ProvisionerUser"
  secret_id   = idsec_sia_secrets_vm.provisioner.secret_id

  provision_format              = var.provision_format
  enable_certificate_validation = var.enable_certificate_validation
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

variable "suffix_pattern" {
  description = "Wildcard pattern for target servers (e.g., '*.web.example.com', '*.prod.local')"
  type        = string
}

variable "account_domain" {
  description = "Domain for the provisioner account"
  type        = string
  default     = "local"
}

variable "secret_name" {
  description = "Name of the strong account"
  type        = string
}

variable "provisioner_username" {
  description = "Username for the RDP account"
  type        = string
}

variable "provisioner_password" {
  description = "Password for the RDP account"
  type        = string
  sensitive   = true
}

variable "target_set_description" {
  description = "Description of the target set"
  type        = string
  default     = "Server group for SIA-RDP access"
}

variable "provision_format" {
  description = "Username format for ephemeral provisioning"
  type        = string
  default     = ""
}

variable "enable_certificate_validation" {
  description = "Enable certificate validation"
  type        = bool
  default     = false
}
```

---

### Target Set for Specific Server

Enables RDP access to a specific server by FQDN or IP address.

main.tf
```terraform
--8<-- "terraform-block.md"

provider "idsec" {
  auth_method = "identity"
  username    = var.idsec_username
  secret      = var.idsec_secret
}

# Strong Account for RDP connections
resource "idsec_sia_secrets_vm" "server_admin" {
  secret_name          = var.secret_name
  secret_type          = "ProvisionerUser"
  provisioner_username = var.provisioner_username
  provisioner_password = var.provisioner_password

  secret_details = jsonencode({
    account_domain = var.account_domain
    description    = "Admin account for ${var.server_address}"
  })
}

# Target Set for Specific Server (FQDN or IP)
resource "idsec_sia_workspaces_target_set" "server" {
  name        = var.server_address
  type        = "Target"
  description = var.target_set_description

  secret_type = "ProvisionerUser"
  secret_id   = idsec_sia_secrets_vm.server_admin.secret_id

  provision_format              = var.provision_format
  enable_certificate_validation = var.enable_certificate_validation
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

variable "server_address" {
  description = "Server FQDN (e.g., 'server01.example.com') or IP address (e.g., '192.168.1.100')"
  type        = string
}

variable "account_domain" {
  description = "Domain for the provisioner account ('local' for standalone servers)"
  type        = string
  default     = "local"
}

variable "secret_name" {
  description = "Name of the strong account"
  type        = string
}

variable "provisioner_username" {
  description = "Username for the RDP account"
  type        = string
}

variable "provisioner_password" {
  description = "Password for the RDP account"
  type        = string
  sensitive   = true
}

variable "target_set_description" {
  description = "Description of the target set"
  type        = string
  default     = "Individual server for SIA-RDP access"
}

variable "provision_format" {
  description = "Username format for ephemeral provisioning"
  type        = string
  default     = ""
}

variable "enable_certificate_validation" {
  description = "Enable certificate validation"
  type        = bool
  default     = false
}
```

---

## Common Patterns

### Environment-Based Segregation

Create separate target sets for different environments:

```terraform
# Development servers
resource "idsec_sia_workspaces_target_set" "dev_servers" {
  name                          = "*.dev.corp.local"
  type                          = "Suffix"
  description                   = "Development environment servers"
  secret_type                   = "ProvisionerUser"
  secret_id                     = idsec_sia_secrets_vm.dev_account.secret_id
  provision_format              = "dev_{guid}"
  enable_certificate_validation = false
}

# Production servers with stricter security
resource "idsec_sia_workspaces_target_set" "prod_servers" {
  name                          = "*.prod.corp.local"
  type                          = "Suffix"
  description                   = "Production environment servers"
  secret_type                   = "PCloudAccount"
  secret_id                     = idsec_sia_secrets_vm.prod_account.secret_id
  provision_format              = "prod_{guid}"
  enable_certificate_validation = true
}
```

### Multiple Target Sets Sharing One Account

A single strong account can be associated with multiple target sets:

```terraform
# Single strong account
resource "idsec_sia_secrets_vm" "shared_admin" {
  secret_name          = "shared-domain-admin"
  secret_type          = "ProvisionerUser"
  provisioner_username = var.admin_username
  provisioner_password = var.admin_password

  secret_details = jsonencode({
    account_domain = "corp.local"
    description    = "Shared admin for multiple target sets"
  })
}

# Domain target set
resource "idsec_sia_workspaces_target_set" "domain" {
  name        = "corp.local"
  type        = "Domain"
  secret_type = "ProvisionerUser"
  secret_id   = idsec_sia_secrets_vm.shared_admin.secret_id
}

# Web servers target set (same account)
resource "idsec_sia_workspaces_target_set" "web_servers" {
  name        = "*.web.corp.local"
  type        = "Suffix"
  secret_type = "ProvisionerUser"
  secret_id   = idsec_sia_secrets_vm.shared_admin.secret_id
}

# Database server target set (same account)
resource "idsec_sia_workspaces_target_set" "db_server" {
  name        = "sqlserver01.corp.local"
  type        = "Target"
  secret_type = "ProvisionerUser"
  secret_id   = idsec_sia_secrets_vm.shared_admin.secret_id
}
```

---

## Optional Fields Reference

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `description` | string | Human-readable description | `"Production web servers"` |
| `provision_format` | string | Format for ephemeral usernames | `"eph_{guid}"`, `"prod_{guid}"` |
| `enable_certificate_validation` | bool | Validate server certificates | `true` or `false` |

### Provision Format Placeholders

The `provision_format` field supports these placeholders:

| Placeholder | Description |
|-------------|-------------|
| `{guid}` | Unique identifier for the session |

Examples:
- `eph_{guid}` → `eph_a1b2c3d4-e5f6-7890-abcd-ef1234567890`
- `prod_{guid}` → `prod_a1b2c3d4-e5f6-7890-abcd-ef1234567890`
- `web_{guid}` → `web_a1b2c3d4-e5f6-7890-abcd-ef1234567890`

---

## Updating Target Sets

### Update Description

```terraform
resource "idsec_sia_workspaces_target_set" "servers" {
  name        = "corp.local"
  type        = "Domain"
  description = "Updated description with more details"  # Changed
  secret_type = "ProvisionerUser"
  secret_id   = idsec_sia_secrets_vm.admin.secret_id
}
```

### Change Provision Format

```terraform
resource "idsec_sia_workspaces_target_set" "servers" {
  name             = "*.example.com"
  type             = "Suffix"
  secret_type      = "ProvisionerUser"
  secret_id        = idsec_sia_secrets_vm.admin.secret_id
  provision_format = "newformat_{guid}"  # Changed from "oldformat_{guid}"
}
```

### Enable Certificate Validation

```terraform
resource "idsec_sia_workspaces_target_set" "servers" {
  name                          = "secure.example.com"
  type                          = "Target"
  secret_type                   = "ProvisionerUser"
  secret_id                     = idsec_sia_secrets_vm.admin.secret_id
  enable_certificate_validation = true  # Changed from false
}
```

---

## Best Practices

1. **Start broad, refine later**: Begin with Domain target sets, add specific Suffix or Target sets as needed
2. **Use meaningful names**: The `name` field should clearly identify the target scope
3. **Add descriptions**: Document the purpose and scope of each target set
4. **Enable certificate validation in production**: Adds security for production systems
5. **Use PCloudAccount for production**: Leverage Privilege Cloud's password rotation for production credentials
6. **Plan for deletion order**: Target sets must be deleted before their associated secrets

---

## Troubleshooting

### Secret Not Found

**Error**: Target set cannot find associated secret

**Solution**: Ensure the secret is created before the target set. Use `depends_on` if needed:

```terraform
resource "idsec_sia_workspaces_target_set" "servers" {
  name        = "corp.local"
  type        = "Domain"
  secret_type = "ProvisionerUser"
  secret_id   = idsec_sia_secrets_vm.admin.secret_id

  depends_on = [idsec_sia_secrets_vm.admin]
}
```

### Certificate Validation Errors

**Error**: Certificate validation fails on target

**Solution**: Ensure valid certificates are installed on target servers, or set `enable_certificate_validation = false` for testing

---

## Related Resources

- [Creating RDP Strong Accounts](creating_sia_rdp_strong_accounts.md) - Create VM secrets for authentication
- [Managing RDP Target Sets and Secrets](managing_rdp_target_sets_and_secrets.md) - Comprehensive management guide

---

## Additional Information

For more details on the resources used in this workflow:

- **Resource**: [idsec_sia_workspaces_target_set](../resources/sia/sia_workspaces_target_set.md)
- **Resource**: [idsec_sia_secrets_vm](../resources/sia/sia_secrets_vm.md)
- **Data Source**: [idsec_sia_workspaces_target_set](../data-sources/sia/sia_workspaces_target_set.md)

