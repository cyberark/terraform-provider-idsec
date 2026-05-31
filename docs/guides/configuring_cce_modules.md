---
page_title: "Configure Connect Cloud Environment (CCE) Modules"
description: |-
  A step-by-step guide that covers onboarding and managing AWS and Azure cloud accounts using the CCE Terraform modules.
---

# Configure Connect Cloud Environment (CCE) Modules

This step-by-step guide instructs you how to automate the onboarding and management of AWS and Azure cloud accounts using the Identity Security Terraform Provider's CCE modules. These modules streamline the process of connecting your cloud infrastructure to the Identity Security platform.

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
- Service-specific submodules (SIA, SCA, Secrets Hub)

Using CCE Terraform resources directly may result in incomplete configurations or security gaps.

## Prerequisites

Before using the CCE Terraform modules, ensure you have:

1. **Terraform installed** - Version 0.13.x or later ([Installation Guide](https://developer.hashicorp.com/terraform/install))
2. **Cloud provider access**:
   - **For AWS**: AWS CLI installed and IAM permissions to:
     - Create and manage IAM roles and policies (`iam:CreateRole`, `iam:PutRolePolicy`, `iam:AttachRolePolicy`)
     - For organization modules: AWS organization read/write permissions (`organizations:DescribeOrganization`, `organizations:ListAccounts`)
   - For Azure, Azure CLI installed and appropriate RBAC permissions
3. **Identity credentials** - Username and secret for the Identity Security provider
4. **Identity tenant configured** - Your tenant must be set up for CCE onboarding

---

## Step 1: Authenticate with Your Cloud Provider

Before running the Terraform module, you must authenticate with your cloud provider to allow the Terraform cloud provider to create and manage resources.

### For AWS Modules

Authenticate using AWS SSO (recommended) or AWS CLI:

**Option A: AWS SSO Login (Recommended)**

If your organization uses AWS IAM Identity Center (formerly AWS SSO), authenticate with the specific account that you need:

```bash
# Log in with a specific SSO profile
aws sso login --profile your-profile-name
```

If you haven't configured your SSO profile yet, run:
```bash
aws configure sso
```

to configure the following:
- SSO start URL (provided by your AWS administrator)
- SSO region
- SSO registration scopes
- CLI default client Region
- CLI default output format
- CLI profile name

**Option B: AWS CLI with Access Keys**

For IAM user credentials:
```bash
aws configure
```

Enter your:
- AWS Access Key ID
- AWS Secret Access Key
- Default region name
- Default output format

**Verify Authentication:**
```bash
aws sts get-caller-identity
```

This should display your account ID, user ID, and ARN.

**Additional Resources:**
- [AWS SSO Login Documentation](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sso.html)
- [AWS CLI Configure Documentation](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-quickstart.html)

### For Azure Modules

Authenticate using Azure CLI:

**Azure Login:**

```bash
az login
```

This will open a browser window for interactive authentication. Sign in with your Azure credentials.

**For non-interactive environments** (service principals):
```bash
az login --service-principal -u <app-id> -p <password-or-cert> --tenant <tenant-id>
```

**List Available Subscriptions:**
```bash
az account list --output table
```

**Set the Active Subscription:**

If you have multiple subscriptions, set the subscription that you want to use:
```bash
az account set --subscription "your-subscription-id-or-name"
```

**Verify Authentication:**
```bash
az account show
```

This should display your current subscription details including subscription ID, name, and tenant ID.

**Additional Resources:**
- [Azure CLI Login Documentation](https://learn.microsoft.com/en-us/cli/azure/authenticate-azure-cli)
- [Azure CLI Get Started Guide](https://learn.microsoft.com/en-us/cli/azure/get-started-with-azure-cli)

---

## Step 2: Configure Terraform Providers

Create or update your `main.tf` file with the required provider configurations.

### Required Providers Block

Add this block to specify all required providers:

```hcl
terraform {
  required_version = ">= 0.13"
  
  required_providers {
    idsec = {
      source  = "cyberark/idsec"
      version = ">= 0.4"
    }
    
    # For AWS modules - include this block
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0"
    }
    
    # For Azure modules - include these blocks
    azurerm = {
      source  = "hashicorp/azurerm"
      version = ">= 3.0"
    }
    
    azuread = {
      source  = "hashicorp/azuread"
      version = ">= 2.0"
    }
  }
}
```

### Configure the Identity Security Provider

The `idsec` provider connects to the Identity Security platform. 
Configure it with your credentials:

```hcl
provider "idsec" {
  auth_method   = "identity_service_user"
  service_user  = var.idsec_service_user
  service_token = var.idsec_service_token
}
```



### Configure the AWS Provider (for AWS modules)

```hcl
provider "aws" {
  region = "us-east-1"  # Change to your preferred region
  
  # Optional: Use a specific profile if using AWS SSO
  # profile = "your-profile-name"
}
```

**Authentication:** The AWS provider uses your AWS CLI authentication from Step 1.

### Configure the Azure Providers (for Azure modules)

```hcl
provider "azurerm" {
  subscription_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"  # your Azure subscription ID
  features {}
}

```

**Note:** You can find your subscription ID by running `az account show` after authentication (Step 1).

**Authentication:** The Azure providers use your Azure CLI authentication from Step 1.

---

## Step 3: Choose Your CCE Module

Select the appropriate module based on what you want to onboard:

### AWS Modules

| Module | Use Case | Documentation |
|--------|----------|---------------|
| **terraform-aws-cce-organization** | Onboard an entire AWS organization | [Module Docs](https://registry.terraform.io/modules/cyberark/cce-organization/aws) |
| **terraform-aws-cce-organization-add-account** | Add member accounts to an existing organization | [Module Docs](https://registry.terraform.io/modules/cyberark/cce-organization-add-account/aws) \| [GitHub](https://github.com/cyberark/terraform-aws-cce-organization-add-account) |
| **terraform-aws-cce-account** | Onboard a single AWS account | [Module Docs](https://registry.terraform.io/modules/cyberark/cce-account/aws) |

### Azure Modules

| Module | Use Case | Documentation |
|--------|----------|---------------|
| **terraform-azure-cce-entra** | Onboard a Microsoft Entra tenant | [Module Docs](https://registry.terraform.io/modules/cyberark/cce-entra/azure) |
| **terraform-azure-cce-management-group** | Onboard an Azure management group | [Module Docs](https://registry.terraform.io/modules/cyberark/cce-management-group/azure) |
| **terraform-azure-cce-subscription** | Onboard an Azure subscription | [Module Docs](https://registry.terraform.io/modules/cyberark/cce-subscription/azure) |
| **terraform-azure-cce-commons** | Create shared Azure SCA resources | [Module Docs](https://registry.terraform.io/modules/cyberark/cce-commons/azure) |

---

## Step 4: Configure Your CCE Module

Add the selected module to your `main.tf` file. Below are detailed examples for each module.

### AWS Organization Example

Onboard an AWS organization with SIA, SCA, and Secrets Hub enabled:

```hcl
module "cce_aws_organization" {
  source  = "cyberark/cce-organization/aws"  
  
  organization_id       = "o-xxxxxxxxxx"
  management_account_id = "123456789012"
  organization_root_id  = "r-xxxx"
  display_name          = "My Organization"
  
  # Enable Secure Infrastructure Access
  sia = { enable = true }
  
  # Enable Secure Cloud Access with SSO
  sca = { 
    enable     = true
    sso_enable = true
    sso_region = "us-east-1"
  }
  
  # Enable Secrets Hub for centralized secrets management
  secrets_hub = {
    enable                  = true
    secrets_manager_regions = ["us-east-1", "us-west-2"]
  }
}
```

**Required Variables:**
- `organization_id` - Your AWS Organization ID (found in the AWS organizations console, format: `o-xxxxxxxxxx`)
- `management_account_id` - Your AWS management account ID (12-digit number)
- `organization_root_id` - The root ID of your organization (format: `r-xxxx`, found in the AWS organizations console)
- `display_name` - A friendly name for your organization

**Optional Service Integrations:**
- `sia` - Secure Infrastructure Access configuration (set `enable = true` to activate)
- `sca` - Secure Cloud Access configuration:
  - `enable` - Enable SCA (true/false)
  - `sso_enable` - Enable AWS SSO integration (true/false, AWS only)
  - `sso_region` - AWS region where SSO is configured (for example, "us-east-1")
- `secrets_hub` - Secrets Hub configuration:
  - `enable` - Enable Secrets Hub (true/false)
  - `secrets_manager_regions` - List of AWS regions for secrets management (for example, `["us-east-1", "us-west-2"]`)

**Resources:**
- [Module Documentation](https://registry.terraform.io/modules/cyberark/cce-organization/aws/latest)
- [Examples](https://github.com/cyberark/terraform-aws-cce-organization/tree/main/examples)

### AWS Organization Add Account Example

Add a member account to an existing AWS organization (run this on member accounts after deploying the organization module):

```hcl
module "cce_add_account" {
  source  = "cyberark/cce-organization-add-account/aws"
  
  # Organization onboarding ID from the organization module output
  org_onboarding_id = "org-abc123"
  
  # Services to enable (must match organization configuration)
  services = ["sia", "sca", "secrets_hub"]
}
```

**Required Variables:**
- `org_onboarding_id` - The organization onboarding ID from the CCE organization module output (obtained from running the organization module on the management account)

**Optional Variables:**
- `services` - List of services to enable for this account (for example, `["sia", "sca", "secrets_hub"]`). Must match services configured in the organization. If not specified, defaults to all services enabled in the organization.

**Important Notes:**
- This module should be run on **AWS member accounts** after the CCE organization module has been deployed to the management account.
- The `org_onboarding_id` comes from the organization module's output.
- Services specified must match those configured in the organization.

**Resources:**
- [Module Documentation](https://registry.terraform.io/modules/cyberark/cce-organization-add-account/aws/latest)
- [GitHub Repository](https://github.com/cyberark/terraform-aws-cce-organization-add-account)
- [Examples](https://github.com/cyberark/terraform-aws-cce-organization-add-account/tree/main/examples)

### AWS Account Example

Onboard an individual AWS account:

```hcl
module "cce_aws_account" {
  source  = "cyberark/cce-account/aws"
  
  account_id           = "123456789012"
  account_display_name = "Production Account"
  
  # Enable services
  sia = { enable = true }
  sca = { enable = true, sso_enable = false }
  
  # Enable Secrets Hub
  secrets_hub = {
    enable                  = true
    secrets_manager_regions = ["us-east-1"]
  }
}
```

**Required Variables:**
- `account_id` - Your AWS account ID (12-digit number)
- `account_display_name` - A friendly name for this account

**Optional Service Integrations:**
- `sia` - Secure Infrastructure Access configuration (set `enable = true` to activate)
- `sca` - Secure Cloud Access configuration:
  - `enable` - Enable SCA (true/false)
  - `sso_enable` - Enable AWS SSO integration (true/false, typically false for individual accounts)
- `secrets_hub` - Secrets Hub configuration:
  - `enable` - Enable Secrets Hub (true/false)
  - `secrets_manager_regions` - List of AWS regions for secrets management

**Resources:**
- [Module Documentation](https://registry.terraform.io/modules/cyberark/cce-account/aws/latest)
- [Examples](https://github.com/cyberark/terraform-aws-cce-account/tree/main/examples)

### Azure Entra Example

Onboard a Microsoft Entra tenant:

```hcl
module "cce_azure_entra" {
  source  = "cyberark/cce-entra/azure"
  
  entra_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  
  sia = { enable = true }
  sca = { enable = true }
}
```

**Required Variables:**
- `entra_id` - Your Microsoft Entra tenant ID (GUID format). You can find the tenant ID in the Azure Portal under Azure Active Directory > Overview > Tenant ID.

**Optional Service Integrations:**
- `sia` - Secure Infrastructure Access configuration (set `enable = true` to activate)
- `sca` - Secure Cloud Access configuration (set `enable = true` to activate)

**Note:** Secrets Hub is currently available for AWS organization modules only.

**Resources:**
- [Module Documentation](https://registry.terraform.io/modules/cyberark/cce-entra/azure/latest)
- [Examples](https://github.com/cyberark/terraform-azure-cce-entra/tree/main/examples)

### Azure Management Group Example

Onboard an Azure management group:

```hcl
module "cce_azure_management_group" {
  source  = "cyberark/cce-management-group/azure"
  
  entra_id            = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  management_group_id = "my-mg-id"
}
```

**Required Variables:**
- `entra_id` - Your Microsoft Entra tenant ID (GUID format).
- `management_group_id` - The ID of the management group to onboard (You can find the management group ID in the Azure Portal under Management Groups).

**Note:** This module onboards management groups for organizational structure management in Azure.

**Resources:**
- [Module Documentation](https://registry.terraform.io/modules/cyberark/cce-management-group/azure/latest)
- [Examples](https://github.com/cyberark/terraform-azure-cce-management-group/tree/main/examples)

### Azure Subscription Example

Onboard an Azure subscription:

```hcl
module "cce_azure_subscription" {
  source  = "cyberark/cce-subscription/azure"
  
  entra_id          = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  entra_tenant_name = "My Tenant"
  subscription_id   = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  subscription_name = "Production Subscription"
  
  sia = { enable = true }
  sca = { enable = true }
}
```

**Required Variables:**
- `entra_id` - Your Microsoft Entra tenant ID (GUID format)
- `entra_tenant_name` - A friendly name for the Entra tenant
- `subscription_id` - Your Azure subscription ID (GUID format, found by running `az account show`)
- `subscription_name` - A friendly name for this subscription

**Optional Service Integrations:**
- `sia` - Secure Infrastructure Access configuration (set `enable = true` to activate)
- `sca` - Secure Cloud Access configuration (set `enable = true` to activate)

**Resources:**
- [Module Documentation](https://registry.terraform.io/modules/cyberark/cce-subscription/azure/latest)
- [Examples](https://github.com/cyberark/terraform-azure-cce-subscription/tree/main/examples)

### Azure Commons (Shared Resources) Example

Create shared Azure SCA resources that can be reused across multiple scopes:

```hcl
module "cce_azure_shared" {
  source  = "cyberark/cce-commons/azure"
  
  entra_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  
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

**Required Variables:**
- `entra_id` - Your Microsoft Entra tenant ID (GUID format)

**SCA Configuration:**
- `sca.enable` - Enable Secure Cloud Access shared resources
- `sca.parameters` - SCA-specific parameters for shared resources (typically set to `null` for initial setup, populated by subsequent module runs)

**Note:** This module creates shared resources that can be reused across multiple Azure scopes (subscriptions, management groups), reducing duplication and simplifying management.

**Resources:**
- [Module Documentation](https://registry.terraform.io/modules/cyberark/cce-commons/azure/latest)
- [Examples](https://github.com/cyberark/terraform-azure-cce-commons/tree/main/examples)

---

## Step 5: Initialize and Apply Terraform

Once your `main.tf` is configured, run the following commands:

### Initialize Terraform

Download required providers and modules:

```bash
terraform init
```

### Review the Execution Plan

Preview the resources that will be created:

```bash
terraform plan
```

Review the output carefully to ensure the configuration matches your expectations.

### Apply the Configuration

Create the resources:

```bash
terraform apply
```

Type `yes` when prompted to confirm the changes.

### Verify the Deployment

After successful application, verify your resources:
- Check the Terraform output for any exported values
- Verify resources in your Identity portal
- Confirm onboarding in your CyberArk Identity portal

---

## Understanding Service Integrations

All CCE Terraform modules support optional service integrations that extend functionality:

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
  sso_enable = true  # AWS organization only - enables SSO integration
  sso_region = "us-east-1"  # AWS organization only - SSO region
}
```

### Secrets Hub (AWS Only)
Enable centralized secrets management between CyberArk and AWS Secrets Manager:
```hcl
secrets_hub = {
  enable                  = true
  secrets_manager_regions = ["us-east-1", "us-west-2"]
}
```

---

## Complete Configuration Example

Here's a complete `main.tf` example for onboarding an AWS organization with all services:

```hcl
terraform {
  required_version = ">= 0.13"
  
  required_providers {
    idsec = {
      source  = "cyberark/idsec"
      version = ">= 0.4"
    }
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0"
    }
  }
}

# Identity Security Provider Configuration
provider "idsec" {
  auth_method   = "identity_service_user"
  service_user  = var.idsec_service_user
  service_token = var.idsec_service_token
}

# AWS Provider Configuration
provider "aws" {
  region = "us-east-1"
}

# Onboard AWS Organization with All Services
module "cce_aws_organization" {
  source  = "cyberark/cce-organization/aws"
  
  organization_id       = "o-xxxxxxxxxx"
  management_account_id = "123456789012"
  organization_root_id  = "r-xxxx"
  display_name          = "My Organization"
  
  # Enable Secure Infrastructure Access
  sia = { enable = true }
  
  # Enable Secure Cloud Access with SSO
  sca = { 
    enable     = true
    sso_enable = true
    sso_region = "us-east-1"
  }
  
  # Enable Secrets Hub
  secrets_hub = {
    enable                  = true
    secrets_manager_regions = ["us-east-1", "us-west-2"]
  }
}

# Outputs
output "org_onboarding_id" {
  description = "Organization onboarding ID for member accounts"
  value       = module.cce_aws_organization.org_onboarding_id
}

output "secrets_hub_role_arn" {
  description = "Secrets Hub IAM role ARN"
  value       = module.cce_aws_organization.secrets_hub_role_arn
}
```

---

## Troubleshooting

### AWS Authentication Issues

**Problem:** `Error: error configuring Terraform AWS Provider`

**Solution:**
- Ensure you've run `aws sso login` or `aws configure`
- Verify with `aws sts get-caller-identity`
- Check your AWS CLI configuration

### Azure Authentication Issues

**Problem:** `Error: building account`

**Solution:**
- Ensure you've run `az login`
- Verify with `az account show`
- Check that you have the correct subscription selected

### Module Not Found

**Problem:** `Error: Module not installed`

**Solution:**
- Run `terraform init` to download modules
- Check your internet connection
- Verify the module source is correct

---

## Module Architecture

Each CCE module follows this structure:
- **Root module**: Handles main resource onboarding and CCE registration
- **Service submodules** (`modules/`): Configure service-specific resources (SIA, SCA, Secrets Hub, etc.)
- **Examples**: Demonstrate common usage patterns in the module's GitHub repository

For detailed information about each module's architecture, see the module's documentation on the Terraform Registry.

---

## Best Practices

1. **Use modules, not individual resources**: Always use the provided CCE modules rather than calling individual resources directly.
2. **Version pinning**: Pin module versions in production environments using the `version` parameter.
3. **Credential management**: 
   - Never hardcode credentials in your Terraform files
   - Consider using HashiCorp Vault, AWS Secrets Manager, or Azure Key Vault.
4. **Iterative onboarding**: Start with basic configurations, then enable services (SIA, SCA, Secrets Hub) as needed.
5. **Testing**: Always validate configurations in non-production environments first.
6. **State management**: Use remote state backends (S3, Azure Storage) for team collaboration.
7. **Review plans**: Always run `terraform plan` and review changes before applying.
8. **Service alignment**: When enabling Secrets Hub, ensure your tenant has the feature enabled.
9. **Regional planning**: For Secrets Hub, carefully plan which AWS regions you need for secrets management.

---

## Additional Resources

### Identity Security Provider
- [Provider Documentation](https://registry.terraform.io/providers/cyberark/idsec/latest/docs)
- [Authentication Guide](https://registry.terraform.io/providers/cyberark/idsec/latest/docs#authentication)
- [Provider GitHub Repository](https://github.com/cyberark/terraform-provider-idsec)

### CCE Modules - AWS
- [AWS Organization Module](https://registry.terraform.io/modules/cyberark/cce-organization/aws/latest)
- [AWS Organization Add Account Module](https://registry.terraform.io/modules/cyberark/cce-organization-add-account/aws/latest)
- [AWS Account Module](https://registry.terraform.io/modules/cyberark/cce-account/aws/latest)

### CCE Modules - Azure
- [Azure Entra Module](https://registry.terraform.io/modules/cyberark/cce-entra/azure/latest)
- [Azure Management Group Module](https://registry.terraform.io/modules/cyberark/cce-management-group/azure/latest)
- [Azure Subscription Module](https://registry.terraform.io/modules/cyberark/cce-subscription/azure/latest)
- [Azure Commons Module](https://registry.terraform.io/modules/cyberark/cce-commons/azure/latest)

### Cloud Provider Documentation
- [AWS Provider Documentation](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [Azure Provider Documentation](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs)
- [Azure AD Provider Documentation](https://registry.terraform.io/providers/hashicorp/azuread/latest/docs)

### Secrets Hub Resources
- [AWS Secrets Manager Documentation](https://docs.aws.amazon.com/secretsmanager/)

### General Resources
- [Developer Portal](https://docs.cyberark.com/)
- [API Documentation](https://api-docs.cyberark.com/)
- [Community](https://cyberark-customers.force.com/s/)

---

## Support

For issues or questions:
- **GitHub Issues**: Report bugs or request features in the relevant repository
  - [Provider Issues](https://github.com/cyberark/terraform-provider-idsec/issues)
  - Module-specific issues: Check each module's GitHub repository
- **Support**: Contact for enterprise support and assistance
- **Community**: Join discussions and get help from other users
