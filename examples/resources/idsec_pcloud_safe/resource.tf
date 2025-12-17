resource "idsec_pcloud_safe" "example_safe" {
  safe_name                = "example_safe"
  description              = "An example safe"
  number_of_days_retention = 0
}

resource "idsec_pcloud_safe_member" "example_safe_members" {
  for_each = toset(var.example_members)

  safe_id        = idsec_pcloud_safe.example_safe.safe_id
  member_name    = each.value
  member_type    = "User"
  permission_set = "connect_only"
}
