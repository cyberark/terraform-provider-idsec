resource "idsec_identity_role_attributes" "myrole_attributes" {
  role_id = "role-id-123"
  attributes = {
    "Department" = "IT"
    "Location"   = "Tel Aviv"
  }
}
