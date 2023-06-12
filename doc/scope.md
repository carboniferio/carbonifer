# Scope

## Infrastructure as Code

### Terraform

Currently, Carbonifer can read only Terraform files. It has been tested with the following versions:

- 1.4.6
- 1.3.7
- 1.3.6

## Cloud providers

In the current state of Carbonifer CLI, it supports resource types described below.

If not in this list, the resource's carbon emissions will be considered to be Zero and reported as `unsupported`.

Not all resource types need to be supported if their energy use is negligible or if impossible to plan (data transfer)

### GCP

| Resource | Limitations  | Comment |
|---|---|---|
| `google_compute_instance`  | | Custom machine, nested boot disk type and GPU supported |
| `google_compute_instance_group_manager`  | | Count will be the target size. Uses machine specifications from `google_compute_instance_template` |
| `google_compute_region_instance_group_manager`  | | Count will be the target size. Uses machine specifications from `google_compute_instance_template` |
| `google_compute_instance_from_template`  | | Uses machine specs from `google_compute_instance_template` |
| `google_compute_autoscaler`  | Takes an average size  | Will set target size of `google_compute_instance_group_manager` |
| `google_compute_disk`| `size` needs to be set, otherwise get it from image| |
| `google_compute_region_disk` | `size` needs to be set, otherwise get it from image| |
| `google_sql_database_instance`  | | Custom machine also supported |

Data resources:

| Resource | Limitations  | Comment |
|---|---|---|
| `google_compute_image`| `disk_size_gb` can be set, otherwise get it from image only if GCP credentials are provided| |

### AWS

| Resource | Limitations  | Comment |
|---|---|---|
| `aws_instance`| No GPU | |
| `aws_ebs_volume`| if size set, or if snapshot declared as data resource | |

Data resources:

| Resource | Limitations  | Comment |
|---|---|---|
| `aws_ami`| `ebs.volume_size` can be set, otherwise get it from image only if AWS credentials are provided| |
| `aws_ebs_snapshot`| `volume_size` can be set, otherwise get it from image only if AWS credentials are provided| |


_more to be implemented_

### Azure

_to be implemented_
