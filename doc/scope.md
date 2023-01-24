# Scope

## Infrastructure as Code

### Terraform

Currently, Carbonifer can read only Terraform files. It has been tested with the following versions:

- 1.3.7
- 1.3.6

## Cloud providers

In the current state of Carbonifer CLI, it supports resource types described below.

If not in this list, the resource's carbon emissions will be considered to be Zero and reported as `unsupported`.

Not all resource types need to be supported if their energy use is negligible or if impossible to plan (data transfer)

### GCP

| Resource | Limitations  | Comment |
|---|---|---|
| `google_compute_instance`  | GCP not supported yet | Custom machine and nested boot disk type supported |
| `google_compute_disk`| `size` needs to be set, otherwise get it from image| |
| `google_compute_region_disk` | `size` needs to be set, otherwise get it from image| |

Data resources:

| Resource | Limitations  | Comment |
|---|---|---|
| `google_compute_image`| `disk_size_gb` needs can be set, otherwise get it from image only if GCP credentials are provided| |

### AWS

_to be implemented_

### Azure

_to be implemented_