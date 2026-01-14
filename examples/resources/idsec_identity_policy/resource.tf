resource "idsec_identity_policy" "myrole_policy" {
  policy_name       = "myrole_policy"
  policy_status     = "Active"
  auth_profile_name = idsec_identity_auth_profile.myrole_auth_profile.auth_profile_name
  role_names = [
    idsec_identity_role.myrole.role_name
  ]
  settings = {
    "/Core/Authentication/IwaSetKnownEndpoint" : "false",
    "/Core/Authentication/IwaSatisfiesAllMechs" : "false",
    "/Core/Authentication/AllowZso" : "true",
    "/Core/Authentication/ZsoSkipChallenge" : "true",
    "/Core/Authentication/ZsoSetKnownEndpoint" : "false",
    "/Core/Authentication/ZsoSatisfiesAllMechs" : "false",
  }
}
