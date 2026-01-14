---
page_title: "Deploying SSH CA"
description: |-
  Deploy an SSH Certificate Authority (CA) to one or more target machines
---

# Deploying SSH CA

## Motivation
The following workflow describes how to securely deploy the SSH Certificate Authority (CA) to existing target machines using the Idsec Terraform Provider.

The SSH CA enables secure, certificate-based SSH authentication by distributing trusted an SSH CA public key to remote target machines. 

The deployment is done over SSH, and the provider fetches and installs the CA script onto each target machine.

## Workflow
The workflow will:
- Authenticate to CyberArk using a user who is a member of the DpaAdmin role.
- Deploy the SSH CA public key to one or more existing target machines over SSH using Terraform.
  
main.tf
```terraform
--8<-- "terraform-block.md"

provider "idsec" {
  auth_method = "identity"
  username    = var.idsec_username
  secret      = var.idsec_secret
}

resource "idsec_sia_ssh_public_key" "ssh_ca" {
  for_each = toset(var.targets)

  target_machine    = each.value
  username          = var.target_username
  private_key_path  = var.target_private_key_path
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

variable "targets" {
  description = "A list of target IPs or hostnames."
  type        = list(string)
}

variable "target_username" {
  description = "The SSH username on the target machine."
  type        = string
}

variable "target_private_key_path" {
  description = "The path to the private key file used for SSH access to the target machine."
  type        = string
}
```

terraform.tfvars
```terraform
targets = [
"192.168.1.10",
"192.168.1.11",
"192.168.1.12"
]

target_username = "ec2-user"
target_private_key_path = "C:\\Users\\example\\.ssh\\id_rsa"
idsec_username = "admin@example.com"
idsec_secret = "supersecret"
```