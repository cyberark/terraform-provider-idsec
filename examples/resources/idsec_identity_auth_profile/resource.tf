resource "idsec_identity_auth_profile" "myrole_auth_profile" {
  auth_profile_name = "myrole_auth_profile"
  first_challenges = [
    "UP"
  ]
  second_challenges = [
    "SMS",
    "EMAIL"
  ]
}
