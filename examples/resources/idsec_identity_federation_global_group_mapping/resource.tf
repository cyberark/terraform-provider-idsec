# Maps the "admins" group assertion of every federation to the "Federated Admins" group.
resource "idsec_identity_federation_global_group_mapping" "admins" {
  attribute_value = "admins"
  group_name      = "Federated Admins"
}
