resource "aws_launch_template" "my_launch_template" {
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

resource "aws_autoscaling_group" "asg_launch_template" {
  availability_zones = ["eu-west-3a"]
  desired_capacity   = 4
  max_size           = 2
  min_size           = 10

  launch_template {
    id      = aws_launch_template.my_launch_template.id
    version = "$Latest"
  }
}