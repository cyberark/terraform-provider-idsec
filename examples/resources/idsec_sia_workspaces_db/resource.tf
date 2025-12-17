resource "idsec_sia_strong_accounts" "pam_account" {
  store_type   = "pam"
  name         = "MyPAMAccount"
  account_name = var.account_name
  safe         = var.safe_name
}


resource "idsec_sia_workspaces_db" "example_db" {
  name                = "example_mssql_db"
  provider_engine     = "mssql-sh-vm"
  read_write_endpoint = var.address
  secret_id           = idsec_sia_strong_accounts.pam_account.id
}
