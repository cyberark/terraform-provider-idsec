resource "idsec_sechub_sync_policy" "example" {
  name        = "example-sync-policy"
  description = "An example Secrets Hub sync policy"

  source = {
    id = "00000000-0000-0000-0000-000000000001"
  }

  target = {
    id = idsec_sechub_secret_store.example_aws.id
  }

  filter = {
    type = "PAM_SAFE"
    data = {
      safe_name = "example-safe"
    }
  }

  transformation = {
    predefined = "password_only_plain_text"
  }
}
