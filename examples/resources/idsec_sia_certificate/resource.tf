resource "idsec_sia_certificate" "my_certificate" {
  cert_name        = "my_certificate"
  cert_description = "My SIA Certificate"
  file             = "/path/to/certificate.pem"
}
