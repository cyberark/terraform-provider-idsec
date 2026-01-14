resource "idsec_policy_vm" "example_policy" {
  metadata = {
    name        = "example_policy"
    description = "Policy for example virtual machine access",
    status = {
      status = "Active"
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
    fqdnipResource = {
      fqdnRules = [
        {
          operator            = "EXACTLY",
          computernamePattern = "myvm.mydomain.com",
          domain              = "domain.com"
        }
      ],
      ipRules = [
        {
          operator = "EXACTLY",
          ipAddresses = [
            "192.168.12.34"
          ],
          logicalName = "CoolLogicalName"
        }
      ]
    }
  }
  behavior = {
    sshProfile = {
      username = "ssh_user"
    },
    rdpProfile = {
      domainEphemeralUser = {
        assignGroups = [
          "rdp_users"
        ],
        enableEphemeralUserReconnect = false,
        assignDomainGroups = [
          "domain_rdp_users"
        ]
      }
    }
  }
}
