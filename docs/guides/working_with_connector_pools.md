---
page_title: "Working with Connector Pools"
description: |-
  Creates and configures a sample connector pool with two assigned networks and two unique identifiers.
---

# Working with Connector Pools

## Motivation

Connector pools are used to group together SIA connectors with other SIA connectors, or system connectors with other system connectors.

The following workflow describes how to create and configure a sample connector pool with two assigned networks and two unique identifiers.

## Workflow
The workflow does the following:
- Creates and configures two networks.
- Creates a connector pool and assigns the networks to it.
- Assigns unique identifiers to the connector pool.

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

resource "idsec_cmgr_network" "network_1" {
  name = var.first_network_name
}

resource "idsec_cmgr_network" "network_2" {
  name = var.second_network_name
}

resource "idsec_cmgr_pool" "pool" {
  name                  = var.pool_name
  description           = "An example connector pool with two assigned networks and two unique identifiers."
  assigned_network_ids  = [idsec_cmgr_network.network_1.network_id, idsec_cmgr_network.network_2.network_id]
}

resource "idsec_cmgr_pool_identifier" "identifier_1" {
  type    = var.first_identifier_type
  value   = var.first_identifier_value
  pool_id = idsec_cmgr_pool.pool.pool_id
}

resource "idsec_cmgr_pool_identifier" "identifier_2" {
  type    = var.second_identifier_type
  value   = var.second_identifier_value
  pool_id = idsec_cmgr_pool.pool.pool_id
}
```

variables.tf
```terraform
variable "idsec_username" {
  description = "The username for the Idsec provider."
  type        = string
}

variable "idsec_secret" {
  description = "The secret/password for the Idsec provider."
  type        = string
  sensitive   = true
}

variable "first_network_name" {
  description = "The name of the first network that is created."
  type        = string
}

variable "second_network_name" {
  description = "The name of the second network that is created."
  type        = string
}

variable "pool_name" {
  description = "The name of the pool that is created."
  type        = string
}

variable "first_identifier_type" {
  description = "The first identifier type of the pool."
  type        = string
}

variable "first_identifier_value" {
  description = "The first identifier value of the pool."
  type        = string
}

variable "second_identifier_type" {
  description = "The second identifier type of the pool."
  type        = string
}

variable "second_identifier_value" {
  description = "The second identifier value of the pool."
  type        = string
}
```
