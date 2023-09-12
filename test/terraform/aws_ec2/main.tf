data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name = "block-device-mapping.volume-size"
    values = [ "30" ]
  }
}

data "aws_ebs_snapshot" "ebs_snapshot" {
  most_recent = true

  filter {
    name   = "volume-size"
    values = ["60"]
  }
}

data "aws_ebs_snapshot" "ebs_snapshot_bigger" {
  most_recent = true

  filter {
    name   = "volume-size"
    values = ["100"]
  }
}

resource "aws_ebs_volume" "ebs_volume" {
  availability_zone = "eu-west-3a"
  snapshot_id       = data.aws_ebs_snapshot.ebs_snapshot_bigger.id
  type = "gp2"
  tags = {
    Name = "ebs_volume"
  }
}

resource "aws_instance" "foo" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = "m5d.xlarge"

  network_interface {
    network_interface_id = aws_network_interface.foo.id
    device_index         = 0
  }

  ebs_block_device {
    device_name = "/dev/sdh"
    volume_size = 20
    volume_type = "standard"
  }

  ebs_block_device {
    device_name = "/dev/sdj"
    snapshot_id = data.aws_ebs_snapshot.ebs_snapshot.id
  }

  ephemeral_block_device {
    device_name  = "/dev/sdk"
    virtual_name = "ephemeral0"
  }

  ephemeral_block_device {
    device_name  = "/dev/sdl"
    virtual_name = "ephemeral1"
  }
}

resource "aws_volume_attachment" "ebs_att" {
  device_name = "/dev/sdi"
  volume_id   = aws_ebs_volume.ebs_volume.id
  instance_id = aws_instance.foo.id
}