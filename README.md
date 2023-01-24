[![Go](https://github.com/carboniferio/carbonifer/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/carboniferio/carbonifer/actions/workflows/go.yml)

# Carbonifer CLI

Command Line Tool to control carbon emission of your cloud infrastructure.
Reading Terraform files, `carbonifer plan` will estimate future Carbon Emissions of infrastructure and help make the right choices to reduce Carbon footprint.

## Scope

This tool can analyze Infrastructure as Code definitions such as:

- [Terraform](https://www.terraform.io/) files
- (more to come)

It can estimate Carbon Emissions of:

- AWS
  - [ ] EC2
  - [ ] RDS
  - [ ] AutoScaling Group
- GCP
  - [x] Compute Engine
  - [ ] Cloud SQL
  - [ ] Instance Group
- Azure
  - [ ] Virtual Machines
  - [ ] Virtual Machine Scale Set
  - [ ] SQL
  
NB: This list of resources will be extended in the future
A list of supported resource types is available in the [Scope](doc/scope.md) document.

## Plan

`carbonifer plan` will read your Terraform folder and estimates Carbon Emissions.

```bash
$ carbonifer plan

 ---------------------------- ---------------- ------------------ 
  resource type                name             emissions         
 ---------------------------- ---------------- ------------------ 
  google_compute_disk          first             0.0422 gCO2eq/h  
  google_compute_instance      first             0.3231 gCO2eq/h  
  google_compute_instance      second            0.4248 gCO2eq/h  
  google_compute_region_disk   regional-first    0.0844 gCO2eq/h  
  google_compute_network       vpc_network      unsupported       
  google_compute_subnetwork    first            unsupported       
 ---------------------------- ---------------- ------------------ 
                               Total             0.8744 gCO2eq/h  
 ---------------------------- ---------------- ------------------ 

```

The report is customizable (text or json, per hour, month...), cf [Configuration](#configuration)

<details><summary>Example of a JSON report</summary>
<p>

```json
{
  "Info": {
    "UnitTime": "h",
    "UnitWattTime": "Wh",
    "UnitCarbonEmissionsTime": "gCO2eq/h",
    "DateTime": "2023-01-24T15:58:25.720493+01:00"
  },
  "Resources": [
    {
      "Resource": {
        "Identification": {
          "Name": "first",
          "ResourceType": "google_compute_disk",
          "Provider": 2,
          "Region": "europe-west9",
          "SelfLink": ""
        },
        "Specs": {
          "Gpu": 0,
          "HddStorage": "1024",
          "SsdStorage": "0",
          "MemoryMb": 0,
          "VCPUs": 0,
          "CPUType": "",
          "ReplicationFactor": 1
        }
      },
      "Power": "0.715",
      "CarbonEmissions": "0.042185",
      "AverageCPUUsage": "0.5"
    },
    {
      "Resource": {
        "Identification": {
          "Name": "first",
          "ResourceType": "google_compute_instance",
          "Provider": 2,
          "Region": "europe-west9",
          "SelfLink": ""
        },
        "Specs": {
          "Gpu": 0,
          "HddStorage": "0",
          "SsdStorage": "1317",
          "MemoryMb": 2480,
          "VCPUs": 1,
          "CPUType": "",
          "ReplicationFactor": 0
        }
      },
      "Power": "5.4755078125",
      "CarbonEmissions": "0.3230549609",
      "AverageCPUUsage": "0.5"
    },
    {
      "Resource": {
        "Identification": {
          "Name": "second",
          "ResourceType": "google_compute_instance",
          "Provider": 2,
          "Region": "europe-west9",
          "SelfLink": ""
        },
        "Specs": {
          "Gpu": 0,
          "HddStorage": "10",
          "SsdStorage": "0",
          "MemoryMb": 4098,
          "VCPUs": 2,
          "CPUType": "",
          "ReplicationFactor": 0
        }
      },
      "Power": "7.1996246093",
      "CarbonEmissions": "0.4247778519",
      "AverageCPUUsage": "0.5"
    },
    {
      "Resource": {
        "Identification": {
          "Name": "regional-first",
          "ResourceType": "google_compute_region_disk",
          "Provider": 2,
          "Region": "europe-west9",
          "SelfLink": ""
        },
        "Specs": {
          "Gpu": 0,
          "HddStorage": "1024",
          "SsdStorage": "0",
          "MemoryMb": 0,
          "VCPUs": 0,
          "CPUType": "",
          "ReplicationFactor": 2
        }
      },
      "Power": "1.43",
      "CarbonEmissions": "0.08437",
      "AverageCPUUsage": "0.5"
    }
  ],
  "UnsupportedResources": [
    {
      "Identification": {
        "Name": "vpc_network",
        "ResourceType": "google_compute_network",
        "Provider": 2,
        "Region": "",
        "SelfLink": ""
      }
    },
    {
      "Identification": {
        "Name": "first",
        "ResourceType": "google_compute_subnetwork",
        "Provider": 2,
        "Region": "europe-west9",
        "SelfLink": ""
      }
    }
  ],
  "Total": {
    "Power": "14.8201324218",
    "CarbonEmissions": "0.8743878128",
    "ResourcesCount": 6
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
| `out.format` | `-f` `--format` | `text` | `text` or `json`
| `out.file` | `-o` `--output`|  | file to write report to
| `data.path` | `<arg>` |  | path of terraform files to analyse
| `avg_cpu_use` |  | `0.5` | planned [average percentage of CPU used](doc/methodology.md#cpu)
| `log` |  | `warn` | level of logs `info`, `debug`, `warn`, `error`
