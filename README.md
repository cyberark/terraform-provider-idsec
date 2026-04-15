![Terraform Provider Idsec](https://github.com/cyberark/terraform-provider-idsec/blob/main/assets/sdk.png)

<p align="center">
    <a href="https://actions-badge.atrox.dev/cyberark/terraform-provider-idsec/goto?ref=master" alt="Build">
        <img src="https://img.shields.io/endpoint.svg?url=https%3A%2F%2Factions-badge.atrox.dev%2Fcyberark%terraform-provider-idsec%2Fbadge%3Fref%3Dmaster&style=flat" />
    </a>
    <a alt="Go Version">
        <img src="https://img.shields.io/github/go-mod/go-version/cyberark/terraform-provider-idsec" />
    </a>
    <a href="https://github.com/cyberark/terraform-provider-idsec/blob/main/LICENSE.txt" alt="License">
        <img src="https://img.shields.io/github/license/cyberark/terraform-provider-idsec?style=flat" alt="License" />
    </a>
</p>

# Terraform Provider Idsec

CyberArk's Official Terraform Provider for CyberArk. This provider allows you to manage CyberArk resources using Terraform.

## Installation

### Install from Terraform Registry

```hcl
terraform {
  required_providers {
    idsec = {
      source  = "cyberark/idsec"
      version = ">= 0.1"
    }
  }
}
```

### Install from Source

```bash
git clone https://github.com/cyberark/terraform-provider-idsec.git
cd terraform-provider-idsec
make build
```

## Provider Configuration

The provider supports multiple authentication methods to connect to CyberArk services. Choose the method that best fits your use case.

### Authentication Methods

| Method | Description | Use Case |
|--------|-------------|----------|
| `identity` | CyberArk Identity personal user authentication | Interactive users, development |
| `identity_service_user` | CyberArk Identity service user authentication | CI/CD pipelines, automation |
| `pvwa` | Password Vault Web Access authentication | PAM Self-Hosted environments |

---

### 1. Identity Authentication (`identity`)

Use this method for personal user authentication via CyberArk Identity.

#### Configuration

```hcl
provider "idsec" {
  auth_method = "identity"
  username    = var.idsec_username
  secret      = var.idsec_password
}
```

#### Environment Variables

```bash
export DEPLOY_ENV=integration-dev          # Optional: for non-production environments
export IDSEC_AUTH_METHOD=identity           # Authentication method
export IDSEC_USERNAME=user@cyberark.cloud   # Your CyberArk username
export IDSEC_SECRET=your-password           # Your password
```

#### Required Attributes

| Attribute | Description |
|-----------|-------------|
| `username` | Your CyberArk Identity username |
| `secret` | Your password |

---

### 2. PVWA Authentication (`pvwa`)

Use this method to authenticate against a Password Vault Web Access (PVWA) server for PAM Self-Hosted environments.

#### Configuration

```hcl
provider "idsec" {
  auth_method       = "pvwa"
  pvwa_url          = "https://pvwa.example.com"
  pvwa_login_method = "cyberark"  # Options: cyberark, ldap, windows
  username          = var.pvwa_username
  secret            = var.pvwa_password
}
```

#### Environment Variables

```bash
export IDSEC_AUTH_METHOD=pvwa
export IDSEC_PVWA_URL=https://pvwa.example.com
export IDSEC_PVWA_LOGIN_METHOD=cyberark
export IDSEC_USERNAME=vault-admin
export IDSEC_SECRET=your-password
```

#### Required Attributes

| Attribute | Description |
|-----------|-------------|
| `pvwa_url` | The base URL of your PVWA server |
| `username` | Your PVWA username |
| `secret` | Your PVWA password |

#### Optional Attributes

| Attribute | Description | Default |
|-----------|-------------|---------|
| `pvwa_login_method` | The PVWA authentication method (`cyberark`, `ldap`, `windows`) | `cyberark` |

---

### Common Provider Attributes

These attributes are available for all authentication methods:

| Attribute | Description | Default |
|-----------|-------------|---------|
| `cache_authentication` | Cache authentication tokens to avoid repeated logins | `true` |


**Note:** For Identity-based authentication methods, the provider automatically discovers your tenant from your username and environment. No tenant URL configuration is required.

## Example Usage

### Using Identity Authentication

```terraform
terraform {
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

resource "idsec_cmgr_network" "example_network" {
  name = "example_network"
}

resource "idsec_cmgr_pool" "example_pool" {
  name                  = "example_pool"
  description           = "A pool for example resources"
  assigned_network_ids  = [idsec_cmgr_network.example_network.network_id]
}

resource idsec_sia_access_connector "example_connector" {
  connector_type    = "ON-PREMISE"
  connector_os      = "linux"
  connector_pool_id = idsec_cmgr_pool.example_pool.pool_id
  target_machine    = "1.1.1.1"
  username          = "ec2-user"
  private_key_path  = "~/.ssh/key.pem"
}
```

In this example, we create a network, a pool, and a SIA connector using the Idsec Terraform provider. The access connector is configured to be installed on the ec2 machine with the given private key and username.

### Using PVWA Authentication (PAM Self-Hosted)

```terraform
terraform {
  required_providers {
    idsec = {
      source  = "cyberark/idsec"
      version = ">= 0.1"
    }
  }
}

provider "idsec" {
  auth_method       = "pvwa"
  pvwa_url          = var.pvwa_url
  pvwa_login_method = "cyberark"
  username          = var.pvwa_username
  secret            = var.pvwa_password
}
```

In this example, we configure the provider to authenticate using PVWA (Password Vault Web Access) for PAM Self-Hosted environments. The `pvwa_login_method` supports `cyberark`, `ldap`, or `windows` authentication methods.

More examples can be found in the [examples](examples) directory.

Provider Configuration can be found in the [provider](docs/index.md) documentation.

Schemas can be found in the relevant documentation for each resource / data source.

## License

This project is licensed under Apache License 2.0 - see [`LICENSE`](LICENSE.txt) for more details

Copyright (c) 2026 CyberArk Software Ltd. All rights reserved.
