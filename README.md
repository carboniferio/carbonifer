[![Go](https://github.com/carboniferio/carbonifer/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/carboniferio/carbonifer/actions/workflows/go.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/carboniferio/carbonifer.svg)](https://pkg.go.dev/github.com/carboniferio/carbonifer)


![Carbonifer Logo](https://user-images.githubusercontent.com/2562534/215261762-f3efb0a2-813b-43d9-a08c-53cdc8825112.png)

https://carbonifer.io/

Command Line Tool to control carbon emission of your cloud infrastructure.
Reading Terraform files, `carbonifer plan` will estimate future Carbon Emissions of infrastructure and help make the right choices to reduce Carbon footprint.

## Scope

This tool can analyze Infrastructure as Code definitions such as:

- [Terraform](https://www.terraform.io/) files
- (more to come)

It can estimate Carbon Emissions of:

- **Google Cloud Platform**
  - [x] **Compute Engine**
    - [x] Compute Instances (generic and custom machine types, and from template)
    - [x] Disks (boot, persistent and region-persistent, HDD or SSD)
    - [X] Machines with GPUs
    - [x] Cloud SQL
    - [x] Instance Group (including regional and with Autoscaler)
- Amazon Web Services
  - [x] EC2

The following will also be supported soon:

- Amazon Web Services
  - [ ] EBS
  - [ ] RDS
  - [ ] AutoScaling Group
- Azure
  - [ ] Virtual Machines
  - [ ] Virtual Machine Scale Set
  - [ ] SQL
  
NB: This list of resources will be extended in the future
A list of supported resource types is available in the [Scope](doc/scope.md) document.

## Install Carbonifer CLI

```bash
go install github.com/carboniferio/carbonifer@latest
```

Go will automatically install it in your $GOPATH/bin directory which should be in your $PATH.

### Docker alternative

Alternatively, you can use the Docker image:

```bash
git clone https://github.com/carboniferio/carbonifer.git
cd carbonifer
docker build -t carbonifer .
docker run -it --rm -v <your_tf_folder_with_config>:/tmp/ carbonifer ./carbonifer --config=/tmp/config.yaml plan /tmp/
```

## Plan

`carbonifer plan` will read your Terraform folder and estimates Carbon Emissions.

```bash
$ carbonifer plan

 ------------------------------ ---------------- ------- ------------------------ 
  resource type                  name             count   emissions per instance  
 ------------------------------ ---------------- ------- ------------------------ 
  google_sql_database_instance   instance         1        2.1716 gCO2eq/h        
  google_compute_disk            first            1        0.0449 gCO2eq/h        
  google_compute_instance        first            1        43.2803 gCO2eq/h       
  google_compute_instance        second           1        0.4489 gCO2eq/h        
  google_compute_region_disk     regional-first   1        0.0898 gCO2eq/h        
  google_compute_subnetwork      first                    unsupported             
  google_compute_network         vpc_network              unsupported             
 ------------------------------ ---------------- ------- ------------------------ 
                                 Total            5        46.0356 gCO2eq/h       

```

In case instances are in a managed group (GCP managed instance group, AWS autoscaling group...), the instances appear in the group name, with a count > 1 and emissions are shown for 1 instance. Of course, `Total` will sum all instances of the group:

```bash
 --------------------------------------- ------------------ ------- ------------------------ 
  resource type                           name               count   emissions per instance  
 --------------------------------------- ------------------ ------- ------------------------ 
  google_compute_instance_group_manager   my-group-manager   3        0.5568 gCO2eq/h        
  google_compute_network                  vpc_network                unsupported             
  google_compute_subnetwork               first                      unsupported             
 --------------------------------------- ------------------ ------- ------------------------ 
                                          Total              3        1.6704 gCO2eq/h        
 --------------------------------------- ------------------ ------- ------------------------ 
 ```

The report is customizable (text or JSON, per hour, month...), cf [Configuration](#configuration)

<details><summary>Example of a JSON report</summary>
<p>

```json
{
  "Info": {
    "UnitTime": "h",
    "UnitWattTime": "Wh",
    "UnitCarbonEmissionsTime": "gCO2eq/h",
    "DateTime": "2023-02-18T14:52:08.757999+01:00",
    "AverageCPUUsage": 0.5,
    "AverageGPUUsage": 0.5
  },
  "Resources": [
    {
      "Resource": {
        "Identification": {
          "Name": "first",
          "ResourceType": "google_compute_disk",
          "Provider": 2,
          "Region": "europe-west9",
          "Count": 1
        },
        "Specs": {
          "GpuTypes": null,
          "HddStorage": "1024",
          "SsdStorage": "0",
          "MemoryMb": 0,
          "VCPUs": 0,
          "ReplicationFactor": 1
        }
      },
      "PowerPerInstance": "0.76096",
      "CarbonEmissionsPerInstance": "0.04489664",
      "AverageCPUUsage": "0.5",
      "Count": "1"
    },
    {
      "Resource": {
        "Identification": {
          "Name": "first",
          "ResourceType": "google_compute_instance",
          "Provider": 2,
          "Region": "europe-west9",
          "Count": 1
        },
        "Specs": {
          "GpuTypes": [
            "nvidia-tesla-a100",
            "nvidia-tesla-k80",
            "nvidia-tesla-k80"
          ],
          "HddStorage": "0",
          "SsdStorage": "1317",
          "MemoryMb": 87040,
          "VCPUs": 12,
          "CPUType": "",
          "ReplicationFactor": 1
        }
      },
      "PowerPerInstance": "733.5648917187",
      "CarbonEmissionsPerInstance": "43.2803286114",
      "AverageCPUUsage": "0.5",
      "Count": "1"
    },
    {
      "Resource": {
        "Identification": {
          "Name": "second",
          "ResourceType": "google_compute_instance",
          "Provider": 2,
          "Region": "europe-west9",
          "Count": 1
        },
        "Specs": {
          "GpuTypes": null,
          "HddStorage": "10",
          "SsdStorage": "0",
          "MemoryMb": 4098,
          "VCPUs": 2,
          "CPUType": "",
          "ReplicationFactor": 1
        }
      },
      "PowerPerInstance": "7.6091047343",
      "CarbonEmissionsPerInstance": "0.4489371793",
      "AverageCPUUsage": "0.5",
      "Count": "1"
    },
    {
      "Resource": {
        "Identification": {
          "Name": "regional-first",
          "ResourceType": "google_compute_region_disk",
          "Provider": 2,
          "Region": "europe-west9",
          "Count": 1
        },
        "Specs": {
          "GpuTypes": null,
          "HddStorage": "1024",
          "SsdStorage": "0",
          "MemoryMb": 0,
          "VCPUs": 0,
          "CPUType": "",
          "ReplicationFactor": 2
        }
      },
      "PowerPerInstance": "1.52192",
      "CarbonEmissionsPerInstance": "0.08979328",
      "AverageCPUUsage": "0.5",
      "Count": "1"
    },
    {
      "Resource": {
        "Identification": {
          "Name": "instance",
          "ResourceType": "google_sql_database_instance",
          "Provider": 2,
          "Region": "europe-west9",
          "Count": 1
        },
        "Specs": {
          "GpuTypes": null,
          "HddStorage": "0",
          "SsdStorage": "10",
          "MemoryMb": 15360,
          "VCPUs": 4,
          "CPUType": "",
          "ReplicationFactor": 2
        }
      },
      "PowerPerInstance": "36.807506875",
      "CarbonEmissionsPerInstance": "2.1716429056",
      "AverageCPUUsage": "0.5",
      "Count": "1"
    }
  ],
  "UnsupportedResources": [
    {
      "Identification": {
        "Name": "vpc_network",
        "ResourceType": "google_compute_network",
        "Provider": 2,
        "Region": "",
        "Count": 1
      }
    },
    {
      "Identification": {
        "Name": "first",
        "ResourceType": "google_compute_subnetwork",
        "Provider": 2,
        "Region": "europe-west9",
        "Count": 1
      }
    }
  ],
  "Total": {
    "Power": "780.264383328",
    "CarbonEmissions": "46.0355986163",
    "ResourcesCount": "5"
  }
}
```

</p>
</details>

## Methodology

This tool will:

1. Read resources from Terraform folder
2. Calculate an estimation of power used by those resources in Watt per Hour
3. Translate this electrical power into an estimation of Carbon Emissions depending on

We can estimate the Carbon Emissions of a resource by taking the electric use of a resource (in Watt-hour) and multiplying it by a carbon emission factor.
This carbon emission factor depends on:

- Cloud Provider
- Region of the Data Center
- The energy mix of this region/country
  - Average
  - (future) real-time

Those calculations and estimations are detailed in the [Methodology document](doc/methodology.md).

## Limitations

We are currently supporting only

- resources with a significative power usage (basically anything which has CPU, GPU, memory or disk)
- resources that can be estimated beforehand (we discard for now data transfer)

Because this is just an estimation, the actual power usage and carbon emission should probably differ depending on the actual usage of the resource (CPU %), and actual grid energy mix (could be weather dependent), ... But that should be enough to take decisions about the choice of provider/region, instance type...

See the [Scope](doc/scope.md) document for more details.

## Usage

`carbonifer [path of terraform files]`

The targeted terraform folder is provided as the only argument. By default, it uses the current folder.

### Prerequisites

- Terraform :
  - Carbonifer uses [Terraform](https://www.terraform.io/):
    - `terrafom` executable available in `$PATH`
    - if not existing, it installs it in a temp folder (`.carbonifer`)
  - [versions supported](doc/scope.md#terraform)
- Cloud provider credentials (optional):
  - if not provided, if terraform does not need it, it won't fail
  - if terraform needs it (to read disk image info...), it will get credentials the same way `terraform` gets its credentials
    - [terraform Google provider](https://registry.terraform.io/providers/hashicorp/google/latest/docs/guides/getting_started#adding-credentials)
    - terraform AWS provider
    - terraform Azure provider

### Configuration

| Yaml key  | CLI flag | Default | Description
|---|---|---|---|
| `unit.time` |   | `h` | Time unit: `h` (hour), `m` (month), `y` (year)
| `unit.power` |   | `w` | Power unit: `W` (watt) or `kW`
| `unit.carbon` |   | `g` | Carbon emission in `g` (gram) or `kg`
| `out.format` | `-f <format>` `--format=<format>` | `text` | `text` or `json`
| `out.file` | `-o <filename>` `--output=<filename>`|  | file to write report to. Default is standard output.
| `data.path` | `<arg>` |  | path of carbonifer data files (coefficents...). Default uses embedded files
| `avg_cpu_use` |  | `0.5` | planned [average percentage of CPU used](doc/methodology.md#cpu)
| `log` |  | `warn` | level of logs `info`, `debug`, `warn`, `error`
