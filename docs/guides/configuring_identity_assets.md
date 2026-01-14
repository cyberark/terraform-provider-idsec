---
page_title: "Configuring Identity Assets"
description: |-
  The following workflow describes how to configure identity assets.
---

# Configuring Identity Assets

## Motivation

The following workflow describes how to configure identity assets using the Idsec Terraform Provider. This will allow the users to create identity roles, assign admin rights to them and configure auth profiles / policies.

## Workflow
The workflow will:
- Login to CyberArk with a normal user
- Create identity roles
- Assign admin rights to them
- Create an auth profile and policy with the above roles

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
  auth_method   = "identity_service_user"
  service_user  = var.idsec_service_user
  service_token = var.idsec_service_token
}

resource "random_string" "random_password" {
  length  = 32
  upper   = true
  lower   = true
  numeric = true
  special = true
}

resource "idsec_identity_user" "myuser" {
    username = var.identity_username
    display_name = var.identity_display_name
    email = var.identity_email
    password = random_string.random_password.result
}

resource "idsec_identity_role" "myrole" {
    role_name = var.identity_role_name
    description = var.identity_role_description
}

resource "idsec_identity_role_member" "myrole_member" {
    role_id = idsec_identity_role.myrole.role_id
    member_name = idsec_identity_user.myuser.username
    member_type = "USER" 
}

resource "idsec_identity_role_admin_rights" "myrole_admin_rights" {
    role_id = idsec_identity_role.myrole.role_id
    admin_rights = [
      "ServiceRight/dpaShowTile"
    ]
}

resource "idsec_identity_auth_profile" "myrole_auth_profile" {
    auth_profile_name = "myrole_auth_profile"
    first_challenges = [
      "UP"
    ]
    second_challenges = [
      "SMS",
      "EMAIL"
    ]
}

resource "idsec_identity_policy" "myrole_policy" {
    policy_name = "myrole_policy"
    policy_status = "Active"
    auth_profile_name = idsec_identity_auth_profile.myrole_auth_profile.auth_profile_name
    role_names = [
      idsec_identity_role.myrole.role_name
    ]
    settings = {
      "/Core/Authentication/IwaSetKnownEndpoint":  "false",
			"/Core/Authentication/IwaSatisfiesAllMechs": "false",
			"/Core/Authentication/AllowZso":             "true",
			"/Core/Authentication/ZsoSkipChallenge":     "true",
			"/Core/Authentication/ZsoSetKnownEndpoint":  "false",
			"/Core/Authentication/ZsoSatisfiesAllMechs": "false",
    }
}
```

variables.tf
```terraform
variable "idsec_service_user" {
  description = "Service user for the CyberArk User"
  type        = string
}

variable "idsec_service_token" {
  description = "Service token for the CyberArk User"
  type        = string
  sensitive   = true
}

variable "identity_username" {
  description = "Username of the identity user to create"
  type        = string
}

variable "identity_display_name" {
  description = "Display name of the identity user to create"
  type        = string
}

variable "identity_email" {
  description = "Email of the identity user to create"
  type        = string
}

variable "identity_role_name" {
  description = "Name of the identity role to create"
  type        = string
}

variable "identity_role_description" {
  description = "Description of the identity role to create"
  type        = string
}
```
