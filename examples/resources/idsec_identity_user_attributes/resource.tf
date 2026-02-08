resource "idsec_identity_user_attributes_schema" "myuser_attributes_schema" {
  columns = [
    {
      name = "EmployeeNumber_Attr1"
      type = "Text"
    },
    {
      name = "CostCenter_Attr2"
      type = "Text"
    }
  ]
}

resource "idsec_identity_user_attributes" "myuser_attributes" {
  depends_on = [idsec_identity_user_attributes_schema.myuser_attributes_schema]
  user_id    = "myuser_id"
  attributes = {
    "EmployeeNumber_Attr1" = "12345"
    "CostCenter_Attr2"     = "CC1001"
  }
}
