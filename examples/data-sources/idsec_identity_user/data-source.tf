data "idsec_identity_user" "myuser" {
  username = "myuser@cyberark.cloud.12345"
}

data "idsec_identity_user" "myuser_by_id" {
  user_id = "MY_USER_ID"
}
