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

The provider automatically discovers your tenant based on your credentials. You only need to provide:

- **`auth_method`** - Authentication method (`identity` or `identity_service_user`)
- **`username`** - Your CyberArk username (for `identity` method)
- **`secret`** - Your password (for `identity` method)


**Note:** No subdomain or tenant URL configuration is required. The provider automatically discovers your tenant from your username and environment.


## Example Usage

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

More examples can be found in the [examples](examples) directory.

Provider Configuration can be found in the [provider](docs/index.md) documentation.

Schemas can be found in the relevant documentation for each resource / data source.

## License

This project is licensed under Apache License 2.0 - see [`LICENSE`](LICENSE.txt) for more details

Copyright (c) 2025 CyberArk Software Ltd. All rights reserved.
