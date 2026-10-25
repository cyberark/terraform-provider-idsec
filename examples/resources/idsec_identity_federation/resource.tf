# The identity provider values below are placeholders: replace them with your identity provider's.
# The server fetches oidc_metadata_url and saml_metadata_url when the federation is created and on
# every update, so they must be real URLs the tenant can reach, or the apply fails.

variable "okta_client_secret" {
  type      = string
  sensitive = true
}

variable "adfs_oauth_client_secret" {
  type      = string
  sensitive = true
}

# An external OIDC identity provider. Register oidc_redirect_url and oidc_initiate_login_url
# with the identity provider after the first apply.
resource "idsec_identity_federation" "okta" {
  federation_name = "Okta"
  federation_type = "OIDC"
  domains         = ["example.com"]
  # Your identity provider's discovery document, ending in /.well-known/openid-configuration.
  oidc_metadata_url  = "https://example.okta.com/.well-known/openid-configuration"
  oidc_client_id     = "0oa1b2c3d4e5f6g7h8i9"
  oidc_client_secret = var.okta_client_secret
  oidc_scopes        = ["openid", "profile", "email"]

  # Link federated users to Cloud Directory users by the email claim, create the missing ones, and keep
  # their display name in sync.
  map_users                    = "Required"
  map_user_attribute           = "email"
  map_user_directory_service   = "09B9A9B0-6CE8-465F-AB03-65766D33B05E"
  map_user_create_if_not_found = true
  attribute_mappings = [
    { target = "DisplayName", source = "name" },
  ]

  # Assertions of the groups claim, mapped in order to federated groups.
  group_mappings = [
    { attribute_value = "okta-admins", group_name = "Okta Admins" },
    { attribute_value = "okta-users", group_name = "Okta Users" },
  ]
}

output "okta_redirect_url" {
  value = idsec_identity_federation.okta.oidc_redirect_url
}

# An external SAML 2.0 identity provider read from its metadata URL. Register saml_sp_metadata_url,
# or saml_sp_acs_url and saml_sp_signing_certificate_url, with the identity provider.
resource "idsec_identity_federation" "adfs" {
  federation_name = "ADFS"
  federation_type = "SAML 2.0"
  domains         = ["corp.example.com"]
  # Your identity provider's SAML metadata document.
  saml_metadata_url   = "https://adfs.example.com/FederationMetadata/2007-06/FederationMetadata.xml"
  saml_sign_algorithm = "SHA256"

  # OAuth credentials that let users of this identity provider sign in with a password, through its
  # token URL.
  saml_oauth_client_id        = "identity-password-signin"
  saml_oauth_client_secret    = var.adfs_oauth_client_secret
  saml_oauth_access_token_url = "https://adfs.example.com/adfs/oauth2/token"
}

# A SAML 2.0 identity provider configured without metadata. saml_metadata_xml = file("metadata.xml")
# uploads a metadata document instead.
resource "idsec_identity_federation" "legacy" {
  federation_name  = "Legacy IdP"
  federation_type  = "SAML 2.0"
  domains          = ["legacy.example.com"]
  saml_sign_in_url = "https://idp.example.com/sso"
  saml_logout_url  = "https://idp.example.com/slo"
  # The identity provider's signing certificate, exported from it as PEM or base64, in a file next to
  # this configuration.
  saml_signing_certificate = file("idp-signing.pem")
}

output "adfs_acs_url" {
  value = idsec_identity_federation.adfs.saml_sp_acs_url
}
