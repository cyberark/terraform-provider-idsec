---
page_title: "Create DB Environment"
description: |-
  The following workflow describes how to create a ZSP DB environment.
---

# Create DB Environment

## Motivation

The following workflow describes how to configure a ZSP DB environment using the Idsec Terraform Provider. This will allow the users to perform connections via CyberArk to their target databases with ephemeral users.

## Workflow
The workflow will:
- Login to CyberArk with a normal user
- Create a secret for the database strong user
- Onboard a database
- Create a policy that defines the access

main.tf
```terraform
--8<-- "terraform-block.md"

provider "idsec" {
  auth_method = "identity"
  username    = var.idsec_username
  secret      = var.idsec_secret
}

resource "idsec_sia_db_strong_accounts" "db_account" {
  store_type        = "managed"
  name              = "MSSQL_Enterprise"
  address           = "sqlserver.example.com"
  database          = "EnterpriseDB"
  platform          = "MSSql"
  port              = "1433"
  username          = var.db_username
  password          = var.db_secret
}

resource "idsec_sia_workspaces_db" "db" {
  name                = "mssql_db"
  provider_engine     = "mssql-sh-vm"
  read_write_endpoint = var.db_address
  secret_id           = idsec_sia_db_strong_accounts.db_account.id
}

resource "idsec_policy_db" "policy" {
  metadata = {
    name        = "policy"
    description = "Policy for database access",
    status = {
      status = "Active"
    },
    policy_entitlement = {
      target_category = "DB",
      location_type   = "FQDN/IP"
    },
    policy_tags = [],
    time_zone = "Asia/Jerusalem"
  }
  principals = [
    {
      id   = "DPA_Admin_Role",
      name = "DpaAdmin",
      type = "ROLE"
    }
  ]
  conditions = {
    access_window = {
      days_of_the_week = [1, 2, 3, 4, 5, 6],
      from_hour        = "09:00",
      to_hour          = "17:00"
    },
    max_session_duration = 8
  }
  targets = {
    "FQDN/IP" = {
      instances = [
        {
          instance_name         = idsec_sia_workspaces_db.db.name,
          instance_type         = "MSSQL",
          instance_id           = tostring(idsec_sia_workspaces_db.db.id),
          authentication_method = "ldap_auth",
          ldap_auth_profile = {
            assign_groups = [
              "HR"
            ]
          }
        }
      ]
    }
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

variable "db_username" {
  description = "Username for the db"
  type        = string
}

variable "db_address" {
  description = "Address of the db"
  type        = string
}

variable "db_secret" {
  description = "Password for the db"
  type        = string
  sensitive   = true
}
```
