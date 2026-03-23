data "idsec_identity_webapp_permission" "my_webapp_permission_by_id" {
  webapp_id      = "my_webapp_id"
  principal_type = "User"
  principal      = "user@cyberark.cloud.12345"
}

data "idsec_identity_webapp_permission" "my_webapp_permission_by_name" {
  webapp_name    = "my_webapp_name"
  principal_type = "Role"
  principal_id   = "MyRoleId"
}
