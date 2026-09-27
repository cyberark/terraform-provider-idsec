# Copyright (c) HashiCorp, Inc.

resource "idsec_policy_k8s" "example_policy" {
  metadata = {
    name        = "example_k8s_policy"
    description = "Example policy for Kubernetes cluster access"
    policy_entitlement = {
      target_category = "Clusters"
      location_type   = "AWS"
      policy_type     = "Recurring"
    }
    policy_tags = ["k8s", "example"]
    time_zone   = "Asia/Jerusalem"
  }
  delegation_classification = "Unrestricted"
  connection_method         = "direct"
  principals = [
    {
      id   = "12345-deac-4bd2-1234-d5b3d112345"
      name = "ab_cde@cyberark.cloud.12345"
      type = "USER"
    }
  ]
  conditions = {
    access_window = {
      days_of_the_week = [1, 2, 3, 4, 5]
      from_hour        = "09:00:00"
      to_hour          = "17:00:00"
    }
    max_session_duration = 1
  }
  targets = {
    aws_account_targets = [
      {
        role_id      = "arn:aws:iam::123456789012:role/K8sClusterAdminRole"
        workspace_id = "123456789012"
        cluster_id   = "arn:aws:eks:us-east-1:123456789012:cluster/my-cluster"
        scope        = "cluster"
      }
    ]
  }
}
