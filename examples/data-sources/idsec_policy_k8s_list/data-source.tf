data "idsec_policy_k8s_list" "example" {
  location_type   = ["AWS"]
  target_category = ["Clusters"]
  policy_type     = ["Recurring"]
  status          = ["Active"]
}
