resource "idsec_sia_settings_https_relay" "example" {
  is_https_relay_enabled = true
  relay_host             = "relay.example.com"
  ssh_relay_port         = 2222
}
