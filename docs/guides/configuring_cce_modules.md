---
page_title: "Configuring Connect Cloud Environment (CCE) Modules"
description: |-
  Guide for using the CCE Terraform modules to onboard and manage cloud accounts across AWS and Azure.
---

# Configuring Connect Cloud Environment (CCE) Modules

The Identity Security Terraform Provider includes modules for automating Connect Cloud Environment (CCE) onboarding across AWS and Azure. These modules streamline the process of connecting your cloud infrastructure to CyberArk's identity security platform.

## Overview

CCE modules automate the onboarding of cloud resources including:
- **AWS**: Organizations, accounts, and organizational units
- **Azure**: Entra, management groups, subscriptions, and shared resources

The modules handle the creation of service principals, federated identity credentials, custom roles, and permissions required for CCE integration.

## Important: Use Modules, Not Resources Directly

**⚠️ Recommendation:** Always use the provided CCE **modules** rather than directly calling individual CCE resources. The modules include:
- Pre-configured permissions and roles
- Proper dependency management
- Validated configuration patterns
- Service-specific submodules (SIA, SCA)

Using resources directly may result in incomplete configurations or security gaps.

## Available CCE Modules

### AWS Modules

#### terraform-aws-cce-organization
Onboards an AWS Organization with optional service integrations.

**Example Usage:**
```hcl
module "cce_aws_organization" {
  source = "cyberark/cce-organization/aws"
  
  organization_id       = "o-xxxxxxxxxx"
  management_account_id = "123456789012"
  organization_root_id  = "r-xxxx"
  display_name          = "My Organization"
  
  sia = { enable = true }
  sca = { 
    enable     = true
    sso_enable = true
    sso_region = "us-east-1"
  }
}
```

#### terraform-aws-cce-account
Onboards an individual AWS account.

**Example Usage:**
```hcl
module "cce_aws_account" {
  source = "cyberark/cce-account/aws"
  
  account_id           = "123456789012"
  account_display_name = "Production Account"
  
  sia = { enable = true }
  sca = { enable = true, sso_enable = false }
}
```

### Azure Modules

#### terraform-azure-cce-entra
Onboards an Azure Entra (tenant) with service integrations.

**Example Usage:**
```hcl
module "cce_azure_entra" {
  source = "cyberark/cce-entra/azure"
  
  entra_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  
  sia = { enable = true }
  sca = { enable = true }
}
```

#### terraform-azure-cce-management-group
Onboards an Azure Management Group.

**Example Usage:**
```hcl
module "cce_azure_management_group" {
  source = "cyberark/cce-management-group/azure"
  
  entra_id            = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  management_group_id = "my-mg-id"
}
```

#### terraform-azure-cce-subscription
Onboards an Azure Subscription with service integrations.

**Example Usage:**
```hcl
module "cce_azure_subscription" {
  source = "cyberark/cce-subscription/azure"
  
  entra_id          = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  entra_tenant_name = "My Tenant"
  subscription_id   = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  subscription_name = "Production Subscription"
  
  sia = { enable = true }
  sca = { enable = true }
}
```

#### terraform-azure-cce-commons
Creates shared Azure resources for SCA that can be reused across multiple scopes.

**Example Usage:**
```hcl
module "cce_azure_shared" {
  source = "cyberark/cce-commons/azure"
  
  entra_id   = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  tenant_id  = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  
  identity_issuer           = var.identity_issuer
  identity_user_id          = var.identity_user_id
  identity_audience         = var.identity_audience
  identity_cloud_tenant_num = "433300"
  
  sca = {
    enable = true
    parameters = {
      sca_entra_onboarding        = true
      sca_entra_app_id            = null
      sca_entra_custom_role_id    = null
      sca_entra_wif_username      = null
      sca_resource_app_id         = null
      sca_resource_custom_role_id = null
      sca_resource_wif_username   = null
    }
  }
}
```

## Service Integrations

All CCE modules support optional service integrations:

### Secure Infrastructure Access (SIA)
Enable workforce access to cloud infrastructure:
```hcl
sia = { enable = true }
```

### Secure Cloud Access (SCA)
Enable automated privilege management for cloud resources:
```hcl
sca = { 
  enable     = true
  sso_enable = true  # AWS only - enables SSO integration
  sso_region = "us-east-1"  # AWS only - SSO region
}
```

## Authentication

Configure the Identity Security Provider with your CyberArk credentials. See the [Identity Security Terraform Provider documentation](https://registry.terraform.io/providers/cyberark/idsec/latest/docs) for authentication configuration details.

## Prerequisites

Before using CCE modules, ensure:
1. You have appropriate permissions in your cloud provider (AWS/Azure)
2. The required cloud provider Terraform providers are configured
3. Your CyberArk tenant is configured for CCE onboarding
4. You have obtained CyberArk Identity credentials for the provider

## Module Architecture

Each CCE module follows this structure:
- **Root module**: Handles main resource onboarding and CCE registration
- **Service submodules** (`services_modules/`): Configure service-specific resources (SIA, SCA, etc.)
- **Examples**: Demonstrate common usage patterns

## Best Practices

1. **Use modules**: Always use the provided modules rather than individual resources
2. **Version pinning**: Pin module versions in production environments
3. **Credential management**: Store sensitive credentials securely (e.g., HashiCorp Vault, AWS Secrets Manager)
4. **Iterative onboarding**: Start with basic configurations, then enable services as needed
5. **Testing**: Validate configurations in non-production environments first

## Additional Resources

- [Identity Security Terraform Provider Documentation](https://registry.terraform.io/providers/cyberark/idsec/latest/docs)
- [Module Source Code](https://github.com/cyberark)
- [CyberArk Community](https://cyberark-customers.force.com/s/)

## Support

For issues or questions:
- GitHub Issues: Report bugs or request features
- CyberArk Support: Contact for enterprise support
