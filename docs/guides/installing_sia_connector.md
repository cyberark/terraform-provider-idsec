---
page_title: "Installing a SIA Connector"
description: |-
  Installing a SIA Connector
---

# Motivation

This workflow describes how to install a SIA connector using the Idsec Terraform Provider. The SIA connector enables secure access to target machines as part of the CyberArk Identity Security Platform.

The installation includes creating a network and a pool. These components are prerequisites for the SIA connector installation and define where the connector is associated.

# Workflow
The workflow will:
- Authenticate to CyberArk with a user who is a member of the DpaAdmin role.
- Create an EC2 machine for the SIA connector installation.
- Create a network.
- Create a pool.
- Install the SIA connector on the EC2 machine and associate it with the pool.

main.tf
```terraform
terraform {
  required_version = ">= 0.13"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.0"
    }
    idsec = {
      source  = "cyberark/idsec"
      version = ">= 0.1"
    }
  }
}

provider "aws" {}

provider "idsec" {
  auth_method = "identity"
  username    = var.idsec_username
  secret      = var.idsec_secret
}

resource "aws_instance" "connector" {
  ami                         = var.connector_ami
  instance_type               = var.connector_instance_type
  key_name                    = var.connector_key_name
  subnet_id                   = var.connector_subnet_id
  vpc_security_group_ids      = var.connector_security_group_ids
  associate_public_ip_address = true
  tags = {
    Name = "connector"
  }
}

resource "idsec_cmgr_network" "network" {
  name = "db_network"
}

resource "idsec_cmgr_pool" "pool" {
  name                  = "db_pool"
  description           = "A pool for the database."
  assigned_network_ids  = [idsec_cmgr_network.network.network_id]
}

resource "idsec_cmgr_pool_identifier" "identifier" {
  type    = "GENERAL_FQDN"
  value   = "*.${var.address}"
  pool_id = idsec_cmgr_pool.pool.pool_id
}

resource "idsec_sia_access_connector" "connector" {
  connector_type    = "ON-PREMISE"
  connector_os      = "linux"
  connector_pool_id = idsec_cmgr_pool.pool.pool_id
  target_machine    = aws_instance.connector.public_ip
  username          = var.connector_username
  private_key_path  = var.connector_private_key_path
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

variable "connector_ami" {
  description = "The AMI ID for the connector instance."
  type        = string
}

variable "connector_instance_type" {
  description = "The instance type for the connector instance."
  type        = string
}

variable "connector_subnet_id" {
  description = "The subnet ID for the connector instance."
  type        = string
}

variable "connector_security_group_ids" {
  description = "The security group IDs for the connector instance."
  type        = list(string)
}

variable "connector_key_name" {
  description = "The key name for the connector instance."
  type        = string
}

variable "connector_username" {
  description = "The username for the connector instance."
  type        = string
}

variable "connector_private_key_path" {
  description = "The private key path for the connector instance."
  type        = string
}
```
