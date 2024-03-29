resource "aws_launch_template" "my_launch_template" {
  name_prefix   = "my_launch_template"
  image_id      = "${data.aws_ami.ubuntu.id}"
  instance_type = "m5d.xlarge"

  block_device_mappings {
    device_name  = "/dev/sdb"
    virtual_name = "ephemeral0"
  }

}

resource "aws_launch_template" "my_launch_template_disk_override" {
  name_prefix   = "my_launch_template"
  image_id      = "${data.aws_ami.ubuntu.id}"
  instance_type = "m5d.xlarge"

  block_device_mappings {
    device_name = "/dev/sda1"

    ebs {
      volume_size = 300
      volume_type = "standard"
      delete_on_termination = true
    }
  }

  block_device_mappings {
    device_name  = "/dev/sdb"
    virtual_name = "ephemeral0"
  }

}

resource "aws_instance" "ec2_with_lt" {
    launch_template {
        id      = "${aws_launch_template.my_launch_template.id}"
        version = "$$Latest"
    }
}

resource "aws_instance" "ec2_with_lt_disk_override" {
    launch_template {
        id      = "${aws_launch_template.my_launch_template_disk_override.id}"
        version = "$$Latest"
    }
}