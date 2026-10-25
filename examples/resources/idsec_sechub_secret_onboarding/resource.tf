# PLAINTEXT secret: the entire secret value is treated as a single password string.
# Use VALUE properties to set PAM account fields as static literals.
data "idsec_sechub_onboard_secret" "plaintext_secret" {
  name     = "my-app/my-plaintext-secret"
  store_id = "store-00000000-0000-0000-0000-000000000001"
}

# JSON secret: individual keys are extracted from the secret's JSON value.
# Use KEY_REF properties to map PAM account fields to specific JSON keys.
# Secret value example: {"username": "...", "address": "...", "password": "..."}
data "idsec_sechub_onboard_secret" "json_secret" {
  name     = "my-app/my-json-secret"
  store_id = "store-00000000-0000-0000-0000-000000000001"
}

# TEMPLATE secret: the secret value is a key-value secret evaluated by Secrets Hub.
# Use KEY_REF properties to map PAM account fields to specific keys.
data "idsec_sechub_onboard_secret" "template_secret" {
  name     = "my-app/my-template-secret"
  store_id = "store-00000000-0000-0000-0000-000000000001"
}

# ── PLAINTEXT ─────────────────────────────────────────────────────────────────
# The secret value is a plain string. Secrets Hub syncs it to PAM as the account password.
# PAM account fields (username, address) are provided as static VALUE literals.
resource "idsec_sechub_secret_onboarding" "example_plaintext" {
  source_store_type = "AWS_ASM"
  target_store_type = "PAM_PCLOUD"
  secret_id         = data.idsec_sechub_onboard_secret.plaintext_secret.id
  secret_value_type = "PLAINTEXT"
  safe_name         = "example-safe"

  pam_account = {
    name                         = "example-plaintext-account"
    platform_id                  = "WinDesktopLocal"
    automatic_management_enabled = false
    properties = {
      username = {
        type  = "VALUE"
        value = "app_user"
      }
      address = {
        type  = "VALUE"
        value = "db.example.com"
      }
    }
  }
}

# ── JSON ──────────────────────────────────────────────────────────────────────
# The secret value is a JSON object. Each PAM account field is mapped to a key
# inside the JSON via KEY_REF. The secret must contain all referenced keys.
resource "idsec_sechub_secret_onboarding" "example_json" {
  source_store_type = "AWS_ASM"
  target_store_type = "PAM_PCLOUD"
  secret_id         = data.idsec_sechub_onboard_secret.json_secret.id
  secret_value_type = "JSON"
  safe_name         = "example-safe"

  pam_account = {
    name                         = "example-json-account"
    platform_id                  = "WinDesktopLocal"
    automatic_management_enabled = false
    properties = {
      username = {
        type    = "KEY_REF"
        key_ref = "username"
      }
      address = {
        type    = "KEY_REF"
        key_ref = "address"
      }
      password = {
        type    = "KEY_REF"
        key_ref = "password" # the key name in your secret, not the password itself
      }
    }
  }
}

# ── TEMPLATE ──────────────────────────────────────────────────────────────────
# The secret value is a key-value secret. The template field defines how the
# password is composed from secret keys (e.g. "{password}").
# PAM account fields can be mapped via KEY_REF or set as static VALUE literals.
resource "idsec_sechub_secret_onboarding" "example_template" {
  source_store_type = "AWS_ASM"
  target_store_type = "PAM_PCLOUD"
  secret_id         = data.idsec_sechub_onboard_secret.template_secret.id
  secret_value_type = "TEMPLATE"
  safe_name         = "example-safe"

  pam_account = {
    name                         = "example-template-account"
    platform_id                  = "WinDesktopLocal"
    automatic_management_enabled = false
    template                     = "{password}"
    properties = {
      username = {
        type  = "VALUE"
        value = "app_user"
      }
      address = {
        type  = "VALUE"
        value = "db.example.com"
      }
    }
  }
}
