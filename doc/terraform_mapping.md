# Terraform mapping

When reading a Terraform file, the Carbonifer CLI uses a mapping. In this mapping, we declare JQ filters to query the Terraform file and extract the necessary information.

For instance, to retrieve the machine type of a GCP instance, we employ the filter:

```jq
.values.machine_type
```

This is a complex but powerful mechanism where we can get any information from the terraform json plan. It can be extented by simply:

- adding a new mapping file in `internal/plan/mappings/<provider>/`
- add a test case in `internal/plan/test`

## Mapping mechanism

Carbonifer will run `terraform plan` and get the terraform file in json format. Then it will apply the JQ filter to get the value of the machine type.

On top of JQ filters, there are some other mechanisms implemented:

- variables and placeholder: they are handy to get a value from somewhere else. Typically used to get data from a template or any other referenced resource.
- regex
- reference files: typically AWS or GCP instance type descriptions (vCPUs, memory, etc.) are stored in a json file and referenced by the mapping
- default values: if a value is not found, we can set a default value
- jq custom methods

Each provider will have a folder under `internal/plan/mappings` where its mapping files will be stored.
All mapping files are merged at runtime to create a single mapping file.
There are 2 types of mapping:

- `.general`:  general configuration for the provider
- `.compute_resource`: compute resource mapping specifications

(in the future, more mapping types can be added)

## General configuration

A special `general.yaml` file is used to describe configuration for the current provider:

```yaml
general:
  aws:
    disk_types:
      default: ssd
      types:
        standard: hdd
        gp2: ssd
        gp3: ssd
        ...
    json_data:
      aws_instances : "aws_instances.json"
    ignored_resources: 
      - "aws_vpc"
      - "aws_volume_attachment"
      - "aws_launch_configuration"
```

- `disk_type`: describe the mapping between the provider disk type and the Carbonifer disk type (ssd or hdd)
- `json_data`: is a list of json files that will be loaded and referenced by the mapping (typically instance type descriptions, where for each instance type we have the number of vCPU, memory, etc.). Those files are read from `internal/data/data_`
- `ignored_resources`: list of resources that will be ignored by Carbonifer (any resource for which carbon emission calculation is no relevant)

## Resources

In other yaml files, the structure of the resource mapping is the following:

```yaml
compute_resource:
  <name of resource>: 
    paths: 
      - cbf::all_select("type";  "aws_instance")
    type: resource
    variables: <...>
    properties: <...>
```

- `<name of resource>`: handy name for this resource, typically we use the same as terraform resource type
- `paths`: list of JQ filters to get the resource from the terraform file
- `type`: type of the resource (resource only for now)
- `variables`: (optional) list of variables and how to resolve it (see below)
- `properties`: list of properties and how to resolve it (see below)

## Properties

Each compute resource need to have the following properties for Carbonifer to be able to calculate its carbon emissions:

- `name`: the name of the resource
- `type`: the type of the resource
- `address`: the address of the resource in the terraform file
- `zone`: the zone(s) of the resource
- `region`: the region of the resource
- `vCPU`: the number of vCPU
- `memory`: the amount of memory in GB (value + unit)
- `storage`: list of storage declared in resource
  - `size`: the size of the storage in GB (value + unit)
  - `type`: the type of the storage

A default value can be set for each property in the mapping file.

A property can have:

- `path`: list of jq filters to get the value from resource in terraform file (example `path: .values.machine_type`)
- `default`: default value if the property is not found in the resource
- `property`: name of the property in the resource (optional, if the path returns a map)
- `regex` : the regex to apply
- `reference`: mechanism to get the value from somewhere else (see below)

A regex is defined as:

- `regex`:
  - `regex`: regex to apply to the value
  - `group`: group of the regex to get the value from

A reference can be:

- `json_file`:
  - `file`: name of the json file to use (actual file is set in `general.yaml`)
  - `property`: name of the property in the json file
- `general`s:
  - `general`: name of the general configuration to use (actual configuration is set in `general.yaml`), example `disk_types` to get the value from `general.aws.disk_types` or `general.gcp.disk_types` mappings
- `path`
  - list other path in the terraform plan
  - `property`: name of the property if the path returns a map
- `return_path`: if true, the value will be the full path itself if values are found, otherwise the value will be the value found at the path

## Path

Path are JQ filters relative to the current resource. For example if we are in a `aws_instance` resource, the path will be relative to this resource (example `.values.machine_type`).

Special JQ methods have been added:

- `cbf::all_select(<property>; <value>)`: select all resources where the property has the value. It could be a root resource or any sub child resource. Example: `cbf::all_select("type";  "aws_instance")`

## Variables

In resource `variables` section we can define variables and how to resolve them. For example:

```yaml
variables:
    properties:
        ami:
          - paths:
            - cbf::all_select("values.image_id";  "${this.values.ami}")
            reference:
              return_path: true
        provider_region:
          - paths:
            - '.configuration'
            property: "region"
```

NB: properties is there just to keep it consitant with other resource definition, all variables are defined under `properties` section.

In this example we have 2 variables:

- `ami`: we have the path where the AMI can be found. It takes `this.values.ami` as input, which is the value `ami` of the current resource. It will return the path of the resource where the AMI is found.
- `provider_region`: decribes where in the json plan the default region is set

Each variable can be defined as any other property below.

Anywhere in the yaml mapping file, we can use a placeholder `${<variable>}` to get the value of the variable. For example:

```yaml
some_property_of_resource:
    - paths: '${ami}.values.block_device_mappings[].ebs | select(length > 0)'
```

In this example ${ami} will be replaced by the path of the AMI referenced by the resource.

There are 3 types of placeholders:

- `${this.<jq path>}`: get the value of the jq path relative to the current resource. Example: `${this.values.foo}` will get the value of the `foo` property of the current resource.
- `${key}`: get the value returned by the path above. Used mainly in variable definitions
- `${<variable>}`: get the value of the variable in the resource `variables` mapping section. Example: `${ami}` will get the value of the variable `ami` defined in the resource `variables` mapping section.

## JQ filters

JQ filters have been chosen because they are widely used, well documented and powerful. Carbonifer is not calling jq but use a [library](https://github.com/itchyny/gojq) that mimics jq. They are also easy to read and understand. And more importantly, easy to run direclty against the terraform json plan by using the `jq` command line tool.
