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

resource "idsec_sia_access_connector" "example_connector" {
  connector_type    = "ON-PREMISE"
  connector_os      = "linux"
  connector_pool_id = var.pool_id
  target_machine    = aws_instance.connector.public_ip
  username          = "ec2-user"
  private_key_path  = "~/.ssh/key.pem"
}
