---
page_title: "Install Connector Management connectors"
description: |-
   Install a connector on a target machine using the idsec_cmgr_connector resource, with examples for Linux (SSH) and Windows (WinRM).
---

# Install Connector Management connectors

This topic describes how to install a Connector Management connector on a target machine using the Idsec Terraform Provider.

# Overview

Connector Management is a SaaS service that acts as an interface between Idira backend servers and your organizational environment. It deploys and monitors Idira components on dedicated servers in any cloud or self-hosted environment.

The agent manages component installations, transfers secured messages, invokes jobs, and reports telemetry from self-hosted infrastructure. It supports both Linux and Windows, and each operating system uses a different connection method:
- **Linux** — connects through SSH using a username and a private key.
- **Windows** — connects through WinRM using either HTTPS (default) or HTTP.

# Before you begin

Ensure the following are in place before you install a connector:

1. **Authenticate to Idira using an account that has permission to create connectors:** Configure the provider block with `auth_method = "identity"` and your credentials, as shown in the workflow examples below.
2. **A connector pool exists:** Assign the connector to a connector pool. For more information, see [Working with Connector Pools](working_with_connector_pools.md) for how to create one.
3. **A target machine is running:** For example, this can be an EC2 instance that the Terraform host can reach over the network.
4. **Network and firewall rules allow the connection:**
   - **Linux** — port **22** (SSH) must be open on the target machine.
   - **Windows** — the relevant WinRM port must be open: **5986** for HTTPS (default) or **5985** for HTTP.
5. **Windows-specific requirements:**
   - The **WinRM service** must be enabled and in **listening mode** on the target machine.
   - If using **HTTPS**, the default protocol, the `trust_certificate` attribute defaults to `false`. This means the target machine's certificate must be present in the certificate store of the machine where Terraform runs. If the certificate is stored at a custom path, provide it using the `certificate_path` attribute.

# Install a connector on a Linux machine

The Linux connector connects through SSH. You must provide a `username` and either a `private_key_path` to specify the path to the key file, or `private_key_contents` to provide the key material directly.
main.tf
```terraform
terraform {
  required_version = ">= 0.13"
  required_providers {
    idsec = {
      source  = "cyberark/idsec"
      version = ">= 0.11"
    }
  }
}
provider "idsec" {
  auth_method = "identity"
  username    = var.idsec_username
  secret      = var.idsec_secret
}

resource "idsec_cmgr_connector" "linux_connector" {
  connector_os      = "linux"
  connector_pool_id = var.pool_id
  target_machine    = var.target_machine
  username          = var.ssh_username
  private_key_path  = var.private_key_path
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

variable "pool_id" {
  description = "The connector pool ID to assign the connector to."
  type        = string
}

variable "target_machine" {
  description = "The hostname or IP address of the target Linux machine."
  type        = string
}

variable "ssh_username" {
  description = "The SSH username for the target machine."
  type        = string
}

variable "private_key_path" {
  description = "Path to the SSH private key file used to connect to the target machine."
  type        = string
}
```

# Install a connector on a Windows machine (WinRM over HTTPS)

When `winrm_protocol` is set to `"https"`, the default, WinRM communicates over port *5986*. Because the `trust_certificate` defaults to `false`, the target machine's certificate must already be trusted by the machine running Terraform. To use a certificate stored at a non-default location, set `certificate_path`.

main.tf
```terraform
terraform {
  required_version = ">= 0.13"
  required_providers {
    idsec = {
      source  = "cyberark/idsec"
      version = ">= 0.11"
    }
  }
}

provider "idsec" {
  auth_method = "identity"
  username    = var.idsec_username
  secret      = var.idsec_secret
}

resource "idsec_cmgr_connector" "windows_connector_https" {
  connector_os      = "windows"
  connector_pool_id = var.pool_id
  target_machine    = var.target_machine
  username          = var.windows_username
  password          = var.windows_password
  winrm_protocol    = "https"
  certificate_path  = var.certificate_path
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

variable "pool_id" {
  description = "The connector pool ID to assign the connector to."
  type        = string
}

variable "target_machine" {
  description = "The hostname or IP address of the target Windows machine."
  type        = string
}

variable "windows_username" {
  description = "The username used to connect to the target Windows machine."
  type        = string
}

variable "windows_password" {
  description = "The password used to connect to the target Windows machine."
  type        = string
  sensitive   = true
}

variable "certificate_path" {
  description = "Path to a custom CA certificate for WinRM HTTPS connections. Leave empty to use the default certificate store."
  type        = string
  default     = ""
}
```

# Install a connector on a Windows machine (WinRM over HTTP)

When `winrm_protocol` is set to `"http"`, WinRM communicates over port *5985*. HTTP connections do not require certificate configuration.

main.tf
```terraform
terraform {
  required_version = ">= 0.13"
  required_providers {
    idsec = {
      source  = "cyberark/idsec"
      version = ">= 0.11"
    }
  }
}

provider "idsec" {
  auth_method = "identity"
  username    = var.idsec_username
  secret      = var.idsec_secret
}

resource "idsec_cmgr_connector" "windows_connector_http" {
  connector_os      = "windows"
  connector_pool_id = var.pool_id
  target_machine    = var.target_machine
  username          = var.windows_username
  password          = var.windows_password
  winrm_protocol    = "http"
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

variable "pool_id" {
  description = "The connector pool ID to assign the connector to."
  type        = string
}

variable "target_machine" {
  description = "The hostname or IP address of the target Windows machine."
  type        = string
}

variable "windows_username" {
  description = "The username used to connect to the target Windows machine."
  type        = string
}

variable "windows_password" {
  description = "The password used to connect to the target Windows machine."
  type        = string
  sensitive   = true
}
```
