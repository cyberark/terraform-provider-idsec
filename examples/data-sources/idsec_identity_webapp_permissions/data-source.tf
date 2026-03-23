data "idsec_identity_webapp_permissions" "my_webapp_permissions_by_id" {
  webapp_id = "my_webapp_id"
}

data "idsec_identity_webapp_permissions" "my_webapp_permissions_by_name" {
  webapp_name = "my_webapp_name"
}
