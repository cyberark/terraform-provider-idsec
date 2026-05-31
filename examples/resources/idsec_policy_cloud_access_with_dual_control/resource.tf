# Copyright (c) HashiCorp, Inc.

resource "idsec_policy_cloud_access" "example_policy_with_dual_control" {
  metadata = {
    name        = "example_policy_with_dual_control"
    description = "Example policy for cloud console access with dual control (access approval)",
    status = {
      status = "Active"
    },
    time_frame = {
      from_time = "2025-12-28T00:00:00"
      to_time   = "2026-02-18T00:00:00"
    },
    policy_entitlement = {
      target_category = "Cloud console",
      location_type   = "AWS",
      policy_type     = "Recurring"
    },
    policy_tags = ["test_policy", "dual_control"],
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
      days_of_the_week = [0, 1, 2, 3, 4, 5, 6],
      from_hour        = null,
      to_hour          = null
    },
    max_session_duration = 1,
    access_approval = {
      required = true,
      approvers = [
        {
          id                    = "67890-abcd-1234-5678-ef1234567890",
          name                  = "approver@cyberark.cloud.12345",
          type                  = "USER",
          source_directory_name = "CyberArk_Cloud_Directory",
          source_directory_id   = "08B9A9B0-8CE8-123F-CD03-12345D33B05H"
        }
      ]
    }
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
