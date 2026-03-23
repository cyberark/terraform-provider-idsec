data "idsec_identity_policies_order" "my_policy_order" {}


data "idsec_identity_policies_order" "my_policy_order_specific" {
  policies_order = [
    "policy_id_1",
    "policy_id_2",
    "policy_id_3"
  ]
}
