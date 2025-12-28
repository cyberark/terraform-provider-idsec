---
page_title: "Create Entra ID policy for entire Microsoft Entra ID directory"
description: |-
  The following workflow describes how to create an Entra ID policy.
---

# Workflow

This workflow demonstrates how to:

1. Get the idsec provider using Identity authentication
    - Uses auth_method = "identity"
    - Uses username (service user)
    - Uses secret (service user password)

2. Run a discovery for structural updates to the workspace that was previously onboarded
    - Uses the following details:  cloud provider, organization_id, and account_info

3. Create an Entra ID policy that gives access on the directory level

main.tf

```terraform
terraform {
  required_version = ">= 0.13"
  required_providers {
    idsec = {
      source  = "cyberark/idsec"
      version = "0.1.0"
    }
  }
}
provider "idsec" {
  auth_method = "identity"
  username    = var.idsec_username
  secret      = var.idsec_secret
}
data "idesc_sca_discovery" "discovery_example" {
  csp             = var.csp
  organization_id = var.organization_id
  account_info = {
    id          =  var.id
    new_account =  var.new_account
  }
}
resource "idsec_policy_cloud_access" "example_entra_id_directory_policy" {
  # Remove if no need for discovery 
  depends_on = [data.idesc_sca_discovery.discovery_example]
  
  metadata = {
    name        = var.name
    description = var.description
    status = {
      status            = var.status
      statusCode        = ""
      statusDescription = ""
      link              = ""
    }
    timeFrame = {
      fromTime = "2023-07-05T12:34:56" #var.from_time
      toTime   = "2023-07-06T12:34:56" #var.to_time
    }
    policyEntitlement = {
      targetCategory = "Cloud console" #var.target_category
      locationType   = "Azure" #var.location_type
      policyType     = "Recurring" #var.policy_type
    }
    policyTags = var.policy_tags
    timeZone = var.time_zone
  }
  conditions = {
    accessWindow = {
      daysOfTheWeek = var.days_of_the_week
      fromHour      = var.from_hour
      toHour        = var.to_hour
    }
    maxSessionDuration = var.max_session_duration
  }
  targets = {
    targets = [
      {
      roleId        = "9b895d92-2cd3-44c7-9d02-a6ac2d5ea5c3" #var.role_id
      workspaceId   = "2ca00f05-abc6-11f5-9f0b-6b3f65b8d1b6" #var.workspace_id
      orgId         = "2ca00f05-abc6-11f5-9f0b-6b3f65b8d1b6" #var.org_id
      workspaceType = "directory" #var.workspace_type
      },
      {
        roleId        = "d2562ede-74db-457e-a7b6-544e236ebb61"
        workspaceId   = "2ca00f05-abc6-11f5-9f0b-6b3f65b8d1b6"
        orgId         = "2ca00f05-abc6-11f5-9f0b-6b3f65b8d1b6"
        workspaceType = "directory"
      }
    ]
  }
  principals = [
    {
      id                  = "c2c7bcc6-1234-44e0-8dff-5be222cd37ee" #var.principal_id
      name                = "John@cyberark.cloud.1234" #var.principal_name
      sourceDirectoryName = "CyberArk Cloud Directory" #var.principal_source_directory_name
      sourceDirectoryId   = "08B9A9B0-8CE8-123F-CD03-12345D33B05H" #var.principal_source_directory_id
      type                = "USER" #var.principal_type
    }
  ]
  delegation_classification = var.delegation_classification
}
```

variables.tf

```terraform
variable "idsec_username" {
  description = "Username for the idsec provider"
  type        = string
}
variable "idsec_secret" {
  description = "Secret/password for the idsec provider"
  type        = string
  sensitive   = true
}
variable "csp" {
  description = "The cloud provider that hosts the workspace to discover (AWS | AZURE | GCP)"
  type        = string
}
variable "organization_id" {
  description = "The ID of the organization to discover (AWS - The AWS organization ID | AZURE: Azure tenant ID GCP: Google Cloud organization ID)"
  type        = string
}
variable "id" {
  description = "The ID of the workspace to discover (AWS - AWS account ID | Azure - Management group, subscription, or resource group ID | GCP - Google Cloud project ID )"
  type        = string
}
variable "new_account" {
  description = "Indicates whether the account is new to an organization, and has not yet been discovered"
  type        = bool
  default     = false
}
variable "name" {
  description = "A unique name for the access policy - minLength: 1, maxLength: 200"
  type        = string
  
  validation {
    condition     = length(var.name) >= 0 && length(var.name) <= 200
    error_message = "The name must be between 0 and 99 characters."
  }
}
variable "description" {
  description = "A short description about the policy - optional | maximum 200 characters"
  type        = string
  default     = ""
  validation {
    condition     = length(var.description) == 0 || length(var.description) <= 200
    error_message = "Description must be empty or up to 200 characters."
  }
}
variable "status" {
  description = "The status of the policy. Allowed values: Active, Suspended, Expired, Validating, Error, Warning. Default: Active."
  type        = string
  default     = "Active"
  validation {
    condition     = var.status == "Active" || var.status == "Suspended" || var.status == "Expired" || var.status == "Validating" || var.status == "Error" || var.status == "Warning"
    error_message = "status must be one of: Active, Suspended, Expired, Validating, Error, Warning."
  }
}
variable "from_time" {
  description = "The date the policy becomes active | pattern: yyyy-MM-ddTHH:mm:ss"
  type        = string
  default     = null
  validation {
    condition     = var.from_time == null || can(regex("^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}$", var.from_time))
    error_message = "fromTime must be null or match the pattern yyyy-MM-ddTHH:mm:ss (e.g., 2025-07-05T12:34:56)."
  }
}
variable "to_time" {
  description = "The date the policy expires. | pattern: yyyy-MM-ddTHH:mm:ss"
  type        = string
  default     = null
  validation {
    condition     = var.to_time == null || can(regex("^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}$", var.to_time))
    error_message = "to_time must be null or match the pattern yyyy-MM-ddTHH:mm:ss (e.g., 2026-07-05T12:34:56)."
  }
}
variable "target_category" {
  description = "The category of the target: Cloud access: Cloud console"
  type        = string
}
variable "location_type" {
  description = "The location of the target. Allowed values: AWS, Azure, GCP."
  type        = string
}
variable "policy_type" {
  description = "The type of policy - Recurring or OnDemand. Choices: Recurring, OnDemand"
  type        = string
}
variable "policy_tags" {
  description = "Customized tags to help identify the policy and those similar to it - maximum 20 tags per policy. (e.g. ['test_azure_microsoft_entra_id'])"
  type        = list(string)
  default     = []
  validation {
    condition     = length(var.policy_tags) == 0 || (length(var.policy_tags) <= 20 && alltrue([for tag in var.policy_tags : can(tag)]))
    error_message = "policy_tags must be an empty list or a list of up to 20 strings."
  }
}
variable "time_zone" {
  description = "The time zone identifier - maxLength: 50"
  type        = string
  default     = "GMT"
  validation {
    condition     = length(var.time_zone) == 0 || length(var.time_zone) <= 50
    error_message = "time_zone must be empty or up to 50 characters."
  }
}
variable "days_of_the_week" {
  description = "The days of the week to include in the policy's access window, where Sunday=0, Monday=1,..., Saturday=6."
  type        = list(number)
  default     = [0, 1, 2, 3, 4, 5, 6]
  validation {
    condition     = length(var.days_of_the_week) == 0 || all([for d in var.days_of_the_week : (d >= 0 && d <= 6)])
    error_message = "Each day must be a number between 0 (Sunday) and 6 (Saturday)."
  }
}
variable "from_hour" {
  description = "The start time of the policy's access window (in HH:mm:ss format, e.g. 08:00:00)"
  type        = string
  default     = null
  validation {
    condition     = var.from_hour == null || can(regex("^\\d{2}:\\d{2}:\\d{2}$", var.from_hour))
    error_message = "from_hour must be null or match the format HH:mm:ss (e.g., 08:00:00)."
  }
}
variable "to_hour" {
  description = "The end time of the policy's access window (in HH:mm:ss format, e.g. 09:00:00)"
  type        = string
  default     = null
  validation {
    condition     = var.to_hour == null || can(regex("^\\d{2}:\\d{2}:\\d{2}$", var.to_hour))
    error_message = "to_hour must be null or match the format HH:mm:ss (e.g., 09:00:00)."
  }
}
variable "max_session_duration" {
  description = "The maximum length of time (in hours) a user can remain connected in a single session. Allowed range: 1-12."
  type        = number
  default     = 1
  validation {
    condition     = var.max_session_duration >= 1 && var.max_session_duration <= 12
    error_message = "max_session_duration must be an number between 1 and 12."
  }
}
variable "role_id" {
  description = "The unique identifier assigned to the IAM role in AWS (IAM role ARN). Required."
  type        = string
}
variable "workspace_id" {
  description = "The workspace ID given to the standalone AWS account when it was onboarded to CyberArk. Required."
  type        = string
}
variable "org_id" {
  description = "AWSOrg:Management account ID (required only for AWS IAM Identity Center). Required."
  type        = string
}
variable "workspace_type" {
  description = "The level at which the Google Cloud organization workspace was onboarded to CyberArk"
  type        = string
}
variable "principal_id" {
  description = "The unique identifier of the identity in CyberArk. An identity is a user, group, or role. maxLength: 40. Required."
  type        = string
  validation {
    condition     = length(var.principal_id) > 0 && length(var.principal_id) <= 40
    error_message = "principal_id must be a non-empty string with a maximum length of 40 characters."
  }
}
variable "principal_name" {
  description = "The name of the principal. minLength: 1, maxLength: 512. Allowed pattern: ^[\\w.+\\-@#]+$ (alphanumeric, dot, plus, hyphen, at, hash). Required."
  type        = string
  validation {
    condition     = length(var.principal_name) >= 1 && length(var.principal_name) <= 512 && can(regex("^[\\w.+\\-@#]+$", var.principal_name))
    error_message = "principal_name must be 1-512 characters and match the pattern ^[\\w.+\\-@#]+$."
  }
}
variable "principal_source_directory_name" {
  description = "The name of the directory service. If the type is ROLE, then this field is optional. maxLength: 256. Allowed pattern: ^\\w+$ (alphanumeric and underscore)."
  type        = string
  validation {
    condition     = length(var.principal_source_directory_name) == 0 || (length(var.principal_source_directory_name) <= 256 && can(regex("^\\w+$", var.principal_source_directory_name)))
    error_message = "principal_source_directory_name must be empty or up to 256 characters and match the pattern ^\\w+$."
  }
}
variable "principal_source_directory_id" {
  description = "The unique identifier of the directory service. If the type is ROLE, then this field is optional and may be empty."
  type        = string
  validation {
    condition     = length(var.principal_source_directory_id) == 0 || length(var.principal_source_directory_id) > 0
    error_message = "principal_source_directory_id must be empty or a non-empty string."
  }
}
variable "principal_type" {
  description = "The type of principal. Allowed values: USER, ROLE, GROUP."
  type        = string
}
variable "delegation_classification" {
  description = "Indicates the user rights for the policy. Allowed values: Restricted, Unrestricted. Default: Unrestricted."
  type        = string
  default     = "Unrestricted"
  validation {
    condition     = var.delegation_classification == "Restricted" || var.delegation_classification == "Unrestricted"
    error_message = "delegation_classification must be either 'Restricted' or 'Unrestricted'."
  }
}
```
