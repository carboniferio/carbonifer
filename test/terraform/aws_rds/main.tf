resource "aws_db_instance" "first" {
  storage_type        = "gp2"
  allocated_storage = 300
  engine               = "mysql"
  instance_class       = "db.t2.large"
  multi_az = true
}

data "aws_db_snapshot" "db_snapshot_100" {
  most_recent = true
  snapshot_type = "public"
  include_public = true
  db_snapshot_identifier = "arn:aws:rds:eu-west-3:516473838419:snapshot:infa-1053-sqlsvr" // Refresh this if needed with the bash script

}

resource "aws_db_instance" "second" {
  snapshot_identifier = data.aws_db_snapshot.db_snapshot_100.id
  engine               = "mysql"
  instance_class       = "db.t2.large"
  multi_az = false
}

resource "aws_db_instance" "third" {
  replicate_source_db         = aws_db_instance.first.id
  engine               = "mysql"
  instance_class       = "db.t2.large"
}

