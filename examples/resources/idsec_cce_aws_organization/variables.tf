# Variables for Idsec Provider Authentication
variable "idsec_username" {
  description = "Idsec Identity username"
  type        = string
  sensitive   = true
}

variable "idsec_secret" {
  description = "Idsec Identity secret/password"
  type        = string
  sensitive   = true
}

variable "idsec_subdomain" {
  description = "Idsec tenant subdomain"
  type        = string
}

# Variables for AWS Organization Configuration
variable "organization_root_id" {
  description = "AWS Organization root ID"
  type        = string
}

variable "management_account_id" {
  description = "AWS Management Account ID"
  type        = string
}

variable "organization_id" {
  description = "AWS Organization ID"
  type        = string
}

variable "account_id" {
  description = "AWS Organization ID"
  type        = string
}

variable "region" {
  description = "AWS region for deployment"
  type        = string
}

variable "organization_display_name" {
  description = "Display name for the organization in CCE UI"
  type        = string
}

variable "scan_organization_role_arn" {
  description = "IAM role ARN for scanning the organization"
  type        = string
}

variable "cross_account_role_external_id" {
  description = "External ID for cross-account role assumption"
  type        = string
}

variable "services" {
  description = "List of services to onboard with their resource configurations"
  type = list(object({
    service_name = string
    resources    = map(string)
  }))
}

