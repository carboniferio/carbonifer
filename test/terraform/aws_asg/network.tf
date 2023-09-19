resource "aws_vpc" "my_vpc" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = "tf-example"
  }
}

resource "aws_subnet" "subnet_1" {
  vpc_id            = aws_vpc.my_vpc.id
  cidr_block        = "172.16.10.0/24"
  availability_zone = "us-west-2a"
  tags = {
    Name = "tf-example"
  }
}

resource "aws_network_interface" "network_interface_1" {
  subnet_id   = aws_subnet.subnet_1.id
  private_ips = ["172.16.10.100"]

  tags = {
    Name = "primary_network_interface"
  }
}

resource "aws_subnet" "subnet_2" {
  vpc_id            = aws_vpc.my_vpc.id
  cidr_block        = "172.16.20.0/24"
  availability_zone = "us-west-2a"
  tags = {
    Name = "tf-example"
  }
}

resource "aws_network_interface" "network_interface_2" {
  subnet_id   = aws_subnet.subnet_2.id
  private_ips = ["172.16.20.100"]

  tags = {
    Name = "primary_network_interface"
  }
}