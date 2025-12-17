data "idsec_cce_aws_account" "example" {
  id = "aaaa1111bbbb2222cccc3333dddd4444" # Onboarding id of account
}

# Output the full data object
output "full_account" {
  value = data.idsec_cce_aws_account.example
}

