resource "idsec_identity_role_attributes_schema" "myrole_attributes_schema" {
  columns = [
    {
      name        = "Department"
      type        = "Text"
      description = "The department of the user"
    },
    {
      name        = "Location"
      type        = "Text"
      description = "The location of the user"
    }
  ]
}
