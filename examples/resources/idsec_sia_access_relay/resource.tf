resource "aws_instance" "relay" {
  ami           = "ami-0bb84b8ffd87024d8"
  instance_type = "t2.micro"
  key_name      = "keypair"
  subnet_id     = "subnet-0f7e01b16e2a381c4"
  tags = {
    Name = "example-relay"
  }
}

resource "idsec_sia_access_relay" "example_relay" {
  https_relay_os    = "linux"
  target_machine    = aws_instance.relay.public_ip
  username          = "ec2-user"
  protocol_port_map = { "SSH" = 2222 }
  private_key_path  = "~/.ssh/key.pem"
}