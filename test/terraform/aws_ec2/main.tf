data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name = "block-device-mapping.volume-size"
    values = [ "30" ]
  }
}

resource "aws_instance" "foo" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = "m4.large"

  network_interface {
    network_interface_id = aws_network_interface.foo.id
    device_index         = 0
  }

  ebs_block_device {
    device_name = "/dev/sdh"
    volume_size = 20
    volume_type = "standard"
  }
}