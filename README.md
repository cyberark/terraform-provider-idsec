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

## Features and Services

- [x] Supported Resources
  - [x] SIA VM Secret
  - [x] SIA DB Secret
  - [x] SIA Target Set Workspace
  - [x] SIA DB Workspace
  - [x] SIA Access Connector
  - [x] Connector Manager Pool
  - [x] Connector Manager Pool Identifier
  - [x] Connector Manager Network
  - [x] PCloud Account
  - [x] PCloud Safe
  - [x] PCloud Safe Member
- [x] Supported Data Sources
  - [x] SIA VM Secret
  - [x] SIA DB Secret
  - [x] SIA Target Set Workspace
  - [x] SIA DB Workspace
  - [x] SIA Access Connector
  - [x] Connector Manager Pool
  - [x] Connector Manager Pool Identifier
  - [x] Connector Manager Network
  - [x] PCloud Account
  - [x] PCloud Safe
  - [x] PCloud Safe Member

## TL;DR

## Installation

### Install from Terraform Registry

```hcl
terraform {
  required_providers {
    idsec = {
      source  = "cyberark/idsec"
      version = ">= 1.0"
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

## Example Usage

```terraform
terraform {
  required_providers {
    idsec = {
      source  = "cyberark/idsec"
      version = ">= 1.0"
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

## Immutable Attributes

Resources can mark attributes as immutable to prevent changes after resource creation. Configure immutable attributes in your action definition:

```go
&actions.IdsecServiceTerraformResourceActionDefinition{
    IdsecServiceBaseTerraformActionDefinition: actions.IdsecServiceBaseTerraformActionDefinition{
        // ... other fields
        ImmutableAttributes: []string{"field_name", "another_field"},
    },
}
```

Attempts to modify immutable attributes will result in clear error messages during terraform plan.

## Customizing Documentation Service Names

Service names in the documentation sidebar can be customized by editing `docs/service-names.yaml`.

### How to Add or Modify Service Names

1. Open `docs/service-names.yaml`
2. Add or update a line with the format: `directory-name: "Display Name"`
3. Run `go run tools/gen-nav.go` to regenerate the navigation
4. Review the changes in `mkdocs.yml`
5. Commit both files to git

Services not listed in the YAML file will automatically display their directory name in uppercase.

## Acceptance tests

Refer to the acceptance tests guide for adding or maintaining provider tests: [Acceptance Tests Documentation](internal/acctest/README.md).

<!-- <NG> -->
## Container based development

Refer to the [Local Container Setup instructions file](local-dev.md)
<!-- </NG> -->
## License

This project is licensed under Apache License 2.0 - see [`LICENSE`](LICENSE.txt) for more details

Copyright (c) 2025 CyberArk Software Ltd. All rights reserved.