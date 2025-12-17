---
page_title: "Working with Service User"
description: |-
  Working with Service User
---

# Motivation

The following workflow describes how to work with Service User using the Idsec Terraform Provider. The Service User is a user in CyberArk's Identity Security Platform that is used for automated tasks and integrations.

# Workflow
The workflow will:
- Login to CyberArk with a service user
- Create a SIA database secret
- Onboard a SIA database

# main.tf
```terraform
terraform {
  required_version = ">= 0.13"
  required_providers {
    idsec = {
      source  = "cyberark/idsec"
      version = "0.1.0"
    }
  }
}

provider "idsec" {
  auth_method   = "identity_service_user"
  service_user  = var.idsec_service_user
  service_token = var.idsec_service_token
}

resource "idsec_sia_strong_accounts" "db_account" {
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
  read_write_endpoint = var.address
  secret_id           = idsec_sia_strong_accounts.db_account.id
}
```

variables.tf
```terraform
variable "idsec_service_user" {
  description = "Service username for the Idsec provider"
  type        = string
}

variable "idsec_service_token" {
  description = "Service token for the Idsec provider"
  type        = string
  sensitive   = true
}

variable "db_username" {
  description = "Database strong account username"
  type        = string
}

variable "db_password" {
  description = "Database strong account password"
  type        = string
  sensitive   = true
}

variable "address" {
  description = "Database address"
  type        = string
}
```
