resource "idsec_pcloud_safe_member" "example_member" {
  safe_id        = var.safe_id
  member_name    = var.member_name
  member_type    = "User"
  permission_set = "read_only"
}
