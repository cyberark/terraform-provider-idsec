resource "idsec_uap_sca" "example_policy" {
  metadata = {
    name        = "example_policy"
    description = "Example policy for cloud console/service access",
    status = {
      status = "Active"
    },
    policy_entitlement = {
      target_category = "Cloud console",
      location_type   = "AWS"
    },
    policy_tags = ["test_policy", "example"],
    time_zone   = "Asia/Jerusalem"
  }
  principals = [
    {
      id   = "12345-deac-4bd2-1234-d5b3d112345",
      name = "ab_cde@cyberark.cloud.12345",
      type = "USER"
    }
  ]
  conditions = {
    access_window = {
      days_of_the_week = [1, 2, 3, 4, 5, 6],
      from_hour        = "09:00:00",
      to_hour          = "17:00:00"
    },
    max_session_duration = 1
  }
  targets = {
    aws_account_targets = [
      {
        role_id        = "arn:aws:iam::123456789012:role/FullAccessDeveloper",
        workspace_id   = "123456789012",
        role_name      = "FullAccessDeveloper",
        workspace_name = "workspace-name"
      }
    ]
  }
}