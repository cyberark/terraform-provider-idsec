resource "idsec_cmgr_pool" "example_pool" {
  name                 = "example_db_pool"
  description          = "A pool for example"
  assigned_network_ids = [var.network_id]
}
