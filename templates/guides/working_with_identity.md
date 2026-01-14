---
page_title: "Working with Identity"
description: |-
  Working with Identity
---

# Working with Identity

## Motivation

This workflow describes how to create users, roles and associate role members using the Idsec Terraform Provider.

## Workflow
The workflow will:
- Authenticate to CyberArk with a user who has admin permissions to create users / roles.
- Create an identity user
- Create an identity role
- Asssociate the user to the role

main.tf
```terraform
--8<-- "terraform-block.md"
provider "idsec" {
  auth_method = "identity"
  username    = var.idsec_username
  secret      = var.idsec_secret
}

resource "idsec_identity_user" "myuser" {
    username = "myuser@cyberark.cloud.12345"
    display_name = "My User"
    email = "myuser@example.com"
    password = "MyPassword"
}

resource "idsec_identity_role" "myrole" {
    role_name = "MyRole"
    description = "An example role created"
}

resource "idsec_identity_role_member" "myrole_user_member" {
    role_id = idsec_identity_role.myrole.role_id
    member_name = idsec_identity_user.myuser.username
    member_type = "USER" 
}
```
