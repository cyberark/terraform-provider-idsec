resource "idsec_identity_role_member" "myrole_user_member" {
  role_id     = "ROLE_ID"
  member_name = "USERNAME"
  member_type = "USER"
}

resource "idsec_identity_role_member" "myrole_role_member" {
  role_id     = "ROLE_ID"
  member_name = "ROLE_NAME"
  member_type = "ROLE"
}

resource "idsec_identity_role_member" "myrole_group_member" {
  role_id     = "ROLE_ID"
  member_name = "GROUP_NAME"
  member_type = "GROUP"
}
