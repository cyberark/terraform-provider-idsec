resource "idsec_identity_user" "myuser" {
  username     = "myuser@cyberark.cloud.12345"
  display_name = "MyUser"
  email        = "myuser@example.com"
  password     = "MyPassword"
}

resource "idsec_identity_user" "myserviceuser" {
  username        = "myserviceuser@cyberark.cloud.12345"
  display_name    = "MyServiceUser"
  email           = "myserviceuser@example.com"
  password        = "MyServicePassword"
  is_service_user = true
  is_oauth_client = true
}
