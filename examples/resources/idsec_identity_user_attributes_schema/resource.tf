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
