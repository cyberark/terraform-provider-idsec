data "idsec_identity_role_member" "myrole_user_member" {
  member_name = "myuser@cyberark.cloud.12345"
  member_type = "USER"
  role_id     = "ROLE_ID"
}

data "idsec_identity_role_member" "myrole_role_member" {
  member_name = "MyRole"
  member_type = "ROLE"
  role_id     = "ROLE_ID"
}

data "idsec_identity_role_member" "myrole_group_member" {
  member_name = "MyGroup"
  member_type = "GROUP"
  role_id     = "ROLE_ID"
}
