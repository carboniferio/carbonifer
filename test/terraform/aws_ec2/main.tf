resource "aws_instance" "foo" {
  ami           = "ami-07b92058b98358eda" # Ubuntu 18 us-west-3
  instance_type = "t2.micro"

  network_interface {
    network_interface_id = aws_network_interface.foo.id
    device_index         = 0
  }
}