resource "idsec_identity_policies_order" "my_policy_order" {
  policies_order = [
    "Policy1",
    "Default Policy",
    "Policy2",
  ]
}
