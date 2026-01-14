resource "idsec_identity_role_admin_rights" "myrole_admin_rights" {
  role_id = idsec_identity_role.myrole.role_id
  admin_rights = [
    "ServiceRight/dpaShowTile"
  ]
}
