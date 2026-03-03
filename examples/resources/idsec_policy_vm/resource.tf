resource "idsec_policy_vm" "example_policy" {
  metadata = {
    name        = "example_policy"
    description = "Policy for example virtual machine access",
    status = {
      status = "Active"
    },
    time_frame = {
      from_time = null
      to_time   = null
    },
    policy_entitlement = {
      target_category = "VM",
      location_type   = "FQDN/IP"
    },
    policy_tags = [],
    time_zone   = "Asia/Jerusalem"
  }
  principals = [
    {
      id   = "DPA_Admin_Role",
      name = "DpaAdmin",
      type = "ROLE"
    }
  ]
  conditions = {
    access_window = {
      days_of_the_week = [1, 2, 3, 4, 5, 6],
      from_hour        = "09:00",
      to_hour          = "17:00"
    },
    max_session_duration = 8
  }
  targets = {
    fqdnip_resource = {
      fqdn_rules = [
        {
          operator             = "EXACTLY",
          computername_pattern = "myvm.mydomain.com",
          domain               = "domain.com"
        }
      ],
      ip_rules = [
        {
          operator = "EXACTLY",
          ip_addresses = [
            "192.168.12.34"
          ],
          logical_name = "CoolLogicalName"
        }
      ]
    }
  }
  behavior = {
    ssh_profile = {
      username = "ssh_user"
    },
    rdp_profile = {
      domain_ephemeral_user = {
        assign_groups = [
          "rdp_users"
        ],
        enable_ephemeral_user_reconnect = false,
        assign_domain_groups = [
          "domain_rdp_users"
        ]
      }
    }
  }
}
