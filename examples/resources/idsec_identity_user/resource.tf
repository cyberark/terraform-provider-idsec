resource "idsec_identity_user" "myuser" {
  username     = "myuser@cyberark.cloud.12345"
  display_name = "MyUser"
  email        = "myuser@example.com"
  password     = "MyPassword"
}
