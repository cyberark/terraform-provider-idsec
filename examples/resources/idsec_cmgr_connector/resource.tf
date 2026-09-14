resource "idsec_cmgr_connector" "example-linux" {
  connector_os      = "linux"
  connector_pool_id = var.pool_id
  target_machine    = var.target_machine
  username          = var.username
  private_key_path  = var.private_key_path
}

resource "idsec_cmgr_connector" "example-windows" {
  connector_os      = "windows"
  connector_pool_id = var.pool_id
  target_machine    = var.target_machine
  username          = var.username
  password          = var.password
  winrm_protocol    = "https"
}