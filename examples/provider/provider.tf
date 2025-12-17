terraform {
  required_providers {
    idsec = {
      source  = "cyberark/idsec"
      version = ">= 1.0"
    }
  }
}

provider "idsec" {
  auth_method = "identity"
  username    = var.idsec_username
  secret      = var.idsec_secret
}

resource "idsec_cmgr_network" "example_network" {
  name = "example_network"
}

resource "idsec_cmgr_pool" "example_pool" {
  name                 = "example_pool"
  description          = "A pool for example resources"
  assigned_network_ids = [idsec_cmgr_network.example_network.network_id]
}

resource "idsec_sia_access_connector" "example_connector" {
  connector_type    = "ON-PREMISE"
  connector_os      = "linux"
  connector_pool_id = idsec_cmgr_pool.example_pool.pool_id
  target_machine    = "1.1.1.1"
  username          = "ec2-user"
  private_key_path  = "~/.ssh/key.pem"
}
