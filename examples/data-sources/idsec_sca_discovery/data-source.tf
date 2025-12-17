data "idsec_sca_discovery" "discovery_example" {
  csp             = var.csp
  organization_id = var.organization_id
  account_info = {
    id          = var.account_id
    new_account = var.new_account
  }
}