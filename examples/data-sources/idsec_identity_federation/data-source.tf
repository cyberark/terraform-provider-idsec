data "idsec_identity_federation" "okta" {
  federation_id = "2f1d6c1e-8a3b-4c5d-9e7f-0a1b2c3d4e5f"
}

output "okta_domains" {
  value = data.idsec_identity_federation.okta.domains
}
