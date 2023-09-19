data "aws_ami" "ubuntu" {
	most_recent = true
  
	filter {
	  name = "block-device-mapping.volume-size"
	  values = [ "30" ]
	}
  }
  