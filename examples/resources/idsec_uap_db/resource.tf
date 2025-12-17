resource "idsec_uap_db" "example_policy" {
  metadata = {
    name        = "example_policy"
    description = "Policy for example database access",
    status = {
      status = "Active"
    },
    policy_entitlement = {
      target_category = "DB",
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
    "FQDN/IP" = {
      instances = [
        {
          instance_name         = "example_db_instance",
          instance_type         = "MSSQL",
          instance_id           = "1234",
          authentication_method = "ldap_auth",
          ldap_auth_profile = {
            assign_groups = [
              "HR"
            ]
          }
        }
      ]
    }
  }
}
