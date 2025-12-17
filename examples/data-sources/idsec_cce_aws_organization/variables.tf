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

# Variables for AWS Organization Data Source
variable "organization_id" {
  description = "CCE onboarding ID of the organization to retrieve"
  type        = string
}

