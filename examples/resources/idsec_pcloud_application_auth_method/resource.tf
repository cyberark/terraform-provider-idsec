resource "idsec_pcloud_application_auth_method" "myapp_auth_method" {
  app_id     = "MyApp"
  auth_type  = "hash"
  auth_value = "myhashvalue"
  comment    = "my app auth method"
}
