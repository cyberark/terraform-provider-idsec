resource "idsec_identity_webapp" "my_aws_webapp" {
  template_name               = "Amazon AWS"
  webapp_name                 = "AWS App"
  description                 = "This is my imported AWS app"
  additional_identifier_value = "123456789234"
  user_name_strategy          = "Fixed"
  username                    = "awsuser"
  password                    = "mypass"
  webapp_login_type           = "AuthenticationRule"
  default_auth_profile        = "AlwaysAllowed"
  auth_rules = {
    enabled    = true
    type       = "RowSet"
    unique_key = "Condition"
    value = [
      {
        conditions = [
          {
            op   = "OpInCorpIpRange"
            prop = "IpAddress"
          }
        ]
        profile_id = "13e3bc1a-6ff7-4b7d-ae90-0ed21d3c393e"
      }
    ]
  }
}

resource "idsec_identity_webapp" "my_aws_with_pcloud_webapp" {
  template_name               = "Amazon AWS"
  webapp_name                 = "AWS App"
  description                 = "This is my imported AWS app"
  additional_identifier_value = "123456789234"
  user_name_strategy          = "Fixed"
  safe                        = "mysafe"
  account_name                = "myaccount"
  ext_account_id              = "123_456"
  is_privileged_app           = true
  webapp_login_type           = "AuthenticationRule"
  default_auth_profile        = "AlwaysAllowed"
  auth_rules = {
    enabled    = true
    type       = "RowSet"
    unique_key = "Condition"
    value = [
      {
        conditions = [
          {
            op   = "OpInCorpIpRange"
            prop = "IpAddress"
          }
        ]
        profile_id = "13e3bc1a-6ff7-4b7d-ae90-0ed21d3c393e"
      }
    ]
  }
}

resource "idsec_identity_webapp" "my_oauth_webapp" {
  template_name        = "OAuth2Server"
  webapp_name          = "OAuth App"
  service_name         = "app_id"
  description          = "This is my imported OAuth app"
  webapp_login_type    = "AuthenticationRule"
  default_auth_profile = "AlwaysAllowed"
  auth_rules = {
    enabled    = true
    type       = "RowSet"
    unique_key = "Condition"
    value = [
      {
        conditions = [
          {
            op   = "OpInCorpIpRange"
            prop = "IpAddress"
          }
        ]
        profile_id = "13e3bc1a-6ff7-4b7d-ae90-0ed21d3c393e"
      }
    ]
  }
  oauth_profile = {
    allowed_auth = ["ClientCreds"]
    audience     = "company://audience"
    issuer       = "mycompany.com"
    known_scopes = [
      {
        scope       = "scope1"
        description = "Scope 1"
      }
    ]
    token_type            = "JwtRS256"
    token_lifetime_string = "0.05:00:00"
  }
}
