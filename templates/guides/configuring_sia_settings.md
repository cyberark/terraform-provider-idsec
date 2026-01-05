---
page_title: "Configuring SIA Settings"
description: |-
  Configure SIA settings globally or per setting using the Idsec Terraform Provider
---

# Motivation
This workflow describes how to manage SIA Settings with the Idsec Terraform Provider.  
SIA settings define global secure-access behavior, including session timeouts, MFA caching, and connection policies.

Using Terraform to configure these settings provides consistent, versioned, and auditable management of your SIA environment.

---
# Two Ways to Manage SIA Settings
The Idsec Terraform Provider supports configuring SIA settings using:

### 1. Global Settings Resource
`idsec_sia_settings_settings`  
- Allows updating multiple settings within a single resource.
- Good for bulk configuration.

### 2. Specific Setting Resources
Each setting has a dedicated resource:
- `idsec_sia_settings_certificate_validation`
- `idsec_sia_settings_ssh_mfa_caching`
- `idsec_sia_settings_rdp_token_mfa_caching`
- `idsec_sia_settings_self_hosted_pam`
- `idsec_sia_settings_logon_sequence`
- …and more.
---

# Workflow
The workflow will:
- Authenticate to CyberArk with a user who is a member of the DpaAdmin role.
- Demonstrate how to update SIA settings using both of the following methods:  
   - Global settings update
   - Specific per setting updates

main.tf
```terraform
--8<-- "terraform-block.md"

provider "idsec" {
  auth_method = "identity"
  username    = var.idsec_username
  secret      = var.idsec_secret
}

# Update multiple SIA settings in one resource
resource "idsec_sia_settings_settings" "global" {
  certificate_validation = {
    enabled = true
  }

  ssh_mfa_caching = {
    is_mfa_caching_enabled  = true
    key_expiration_time_sec = 3600
  }

  standing_access = {
    standing_access_available = true
    session_max_duration      = 120
  }
}


# Update Settings Using Specific Resources
resource "idsec_sia_settings_rdp_recording" "recording" {
  enabled = true
}

resource "idsec_sia_settings_ssh_mfa_caching" "ssh_mfa" {
  is_mfa_caching_enabled  = true
  key_expiration_time_sec = 3600
  client_ip_enforced      = false
}

```

variables.tf
```terraform
variable "idsec_username" {
  description = "The username for the Idsec provider."
  type        = string
}

variable "idsec_secret" {
  description = "The Secret/password for the Idsec provider."
  type        = string
  sensitive   = true
}

```
