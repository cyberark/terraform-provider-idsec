resource "idsec_pcloud_account" "example_account" {
  name        = "example_account"
  platform_id = "WinDesktopLocal"
  username    = var.username
  address     = var.address
  secret      = var.secret
  safe_name   = var.safe_name
}
