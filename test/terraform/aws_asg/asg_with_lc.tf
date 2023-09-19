
resource "aws_autoscaling_group" "asg_with_launchconfig" {
  name = "asg_with_launchconfig"
  max_size = 10
  min_size = 2
  desired_capacity = 4
  vpc_zone_identifier = [aws_subnet.subnet_1.id, aws_subnet.subnet_2.id]
  launch_configuration = aws_launch_configuration.asg_launch_config.name
  
}

resource "aws_launch_configuration" "asg_launch_config" {
  name = "asg_launch_config"
  image_id = data.aws_ami.ubuntu.id
  instance_type = "m5d.xlarge"

  ephemeral_block_device {
    device_name  = "/dev/sdk"
    virtual_name = "ephemeral0"
  }
}