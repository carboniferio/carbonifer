[![Go](https://github.com/carboniferio/carbonifer/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/carboniferio/carbonifer/actions/workflows/go.yml)
# Carbonifer CLI

Command Line Tool to control carbon emission of your cloud infrastructure

## Scope

This tool can analyze Infrastructure as Code definitions such as:

- [Terraform](https://www.terraform.io/) files
- (more to come)

It can estimate carbon emissions of:

- AWS
  - [ ] EC2
  - [ ] RDS
  - [ ] AutoScaling Group
- GCP
  - [ ] Compute Engine
  - [ ] Cloud SQL
  - [ ] Instance Group
- Azure
  - [ ] Virtual Machines
  - [ ] Virtual Machine Scale Set
  - [ ] SQL
  
NB: This list of resources will be extended in the future

## Plan

`carbonifer plan` will read your Terraform folder and estimates carbon emissions.
