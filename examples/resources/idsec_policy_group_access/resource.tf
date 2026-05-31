# Copyright (c) HashiCorp, Inc.

resource "idsec_policy_group_access" "example_policy" {
  metadata = {
    name        = "example_policy"
    description = "Example policy for Groups",
    status = {
      status = "Active"
    },
    time_frame = {
      from_time = "2026-01-28T00:00:00"
      to_time   = "2026-10-18T00:00:00"
    },
    policy_entitlement = {
      target_category = "Groups"
      location_type   = "Azure"
      policy_type     = "Recurring"
    },
    policy_tags = ["test_policy", "example"],
    time_zone   = "Asia/Jerusalem"
  }
  delegation_classification = "Unrestricted"
  principals = [
    {
      id   = "12345-deac-4bd2-1234-d5b3d112345",
      name = "ab_cde@cyberark.cloud.12345",
      type = "USER"
    }
  ]
  conditions = {
    access_window = {
      time_zone        = "Asia/Jerusalem"
      days_of_the_week = [1, 2, 3, 4, 5, 6],
      from_hour        = "09:00:00",
      to_hour          = "17:00:00"
    },
    max_session_duration = 1
  }
  targets = {
    targets = [
      {
        group_id     = "01234567-0abc-1def-2abc-3def45678901"
        directory_id = "89abcdef-4abc-5def-6abc-7def89abcdef"
        group_name   = "group_name"
        group_type   = "security"
      }
    ]
  }
}
