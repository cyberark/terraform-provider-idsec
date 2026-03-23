resource "idsec_identity_webapp_permission" "my_webapp_permission" {
  webapp_id      = "my_app_id"
  principal      = "user@cyberark.cloud.12345"
  principal_type = "User"
  rights = [
    "Admin",
    "Grant",
    "View",
    "ViewDetail",
    "Execute",
    "Automatic",
    "Delete"
  ]
}
