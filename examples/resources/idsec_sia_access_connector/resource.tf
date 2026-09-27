resource "aws_instance" "connector" {
  ami                         = "ami-0bb84b8ffd87024d8"
  instance_type               = "t2.micro"
  key_name                    = "keypair"
  subnet_id                   = "subnet-0f7e01b16e2a381c4"
  vpc_security_group_ids      = ["sg-0d7ca71b7eb03bfbb"]
  associate_public_ip_address = true
  tags = {
    Name = "example-connector"
  }
}

# Machine-based connector (ON-PREMISE, AWS, AZURE, or GCP)
resource "idsec_sia_access_connector" "example_connector" {
  connector_type    = "ON-PREMISE"
  connector_os      = "linux"
  connector_pool_id = var.pool_id
  target_machine    = aws_instance.connector.public_ip
  username          = "ec2-user"
  private_key_path  = "~/.ssh/key.pem"
}

# Kubernetes ephemeral connector — no target_machine or username required
resource "idsec_sia_access_connector" "example_k8s_connector" {
  connector_os      = "k8s-ephemeral"
  connector_pool_id = var.pool_id

  # Required so the connector can be uninstalled later (a helm release may back
  # multiple replica connectors with no single connector_id).
  k_8_s_namespace = "cyberark-sia"

  k_8_s_details = {
    k_8_s_namespace         = "cyberark-sia"
    k_8_s_image_uri         = "registry.example.com/sia-connector:latest"
    k_8_s_username          = "svc-connector"
    k_8_s_password          = var.k8s_password
    k_8_s_replicas          = 2
    k_8_s_image_pull_secret = "sia-pull-secret"
  }
}
