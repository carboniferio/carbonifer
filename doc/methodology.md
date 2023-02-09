# Methodology

We are largely using the work of [Carbon Footprint Calculator](https://www.cloudcarbonfootprint.org/docs/methodology/) and [Esty Cloud Jewels](https://www.etsy.com/codeascraft/cloud-jewels-estimating-kwh-in-the-cloud/) for Energy and Carbon Emission estimations.
Those are just "estimations" and will probably differ from the actual energy used and carbon emissions of your infrastructure.

In summary, for each resource, Carbonifer calculate an [Energy Estimate](#energy-estimate) (Watt per Hour) used by it, and multiply it by the [Carbon Intensity](#carbon-intensity) of the underlying data center.

```text
Estimated Carbon Emissions (gCO2eq/h) = Energy Estimate (Wh) x Carbon Intensity (gCO2eq/Wh)
```

## Energy Estimate

Providers don't really communicate the Energy used by their resources. So we need to come up with an estimation based on different benchmarks and empirical tests. Also, because we do this estimation **before** deploying the infrastructure, at "plan" phase, we need to assume the average use of those resources.

### CPU

To calculate an average use of a CPU, we need the minimum and max power of this CPU and use the following formula:

```text
Average Watts = Number of vCPU * (Min Watts + Avg vCPU Utilization * (Max Watts - Min Watts))
```

- `Average Watts` result in Watt Hour
- `Number of vCPU` : depends on the machine type chosen
  - [GCP machine types](../data/gcp_instances.json) 
  - AWS
  - Azure
- `Min Watt` and `Max Watts` depend on CPU architecture
  - If processor architecture is unknown, we use averages computed by [Carbon Footprint Calculator](https://www.cloudcarbonfootprint.org/docs/methodology/#appendix-i-energy-coefficients): [energy coefficients](../data/energy_coefficients.json)
  - If we do know them, we use a more detailed list:
    - [GCP Watt per CPU type](../data/gcp_watt_cpu.csv)
- `Avg vCPU Utilization` because we do this estimation at "plan" time, there is no way to pick a relevant value. However, to be able to plan and compare different CPUs or regions we need to set this constant. This is read from (by descending priority order)
  - user's config file in `$HOME/.carbonifer/config.yml`), variable `avg_cpu_use`
  - targeted folder config file in `$TERRAFORM_PROJECT/.carbonifer/config.yml`), variable `avg_cpu_use`
  - The default is `0.5` (50%)

### Memory

Using the same methodology of [Carbon Footprint Calculator](https://www.cloudcarbonfootprint.org/docs/methodology/#memory) we also pick the Energy Coefficient of `0.392 Watt Hour / Gigabyte` and we use the following formula:

```text
Watt hours = Memory usage (GB) x Memory Energy Coefficient
```

### Disk Storage

We are using the same `Storage Energy Coefficient` as [Carbon Footprint Calculator](https://www.cloudcarbonfootprint.org/docs/methodology/#storage) in [energy coefficients file](../data/energy_coefficients.json). This coefficient is different for SSD and HDD, so disk type is important.

```text
Watt hours = Disk Size (TB) x Storage Energy Coefficient x Replication Factor
```

``

`Replication Factor`: most cloud provider offers to the customer data replication to minimize the risk of data loss:

- Regular Disk will have a replication factor of 1
- GCP regional disk: user sets the list of zones to replicate data, so the Replication Factor will be equal to the number of zones picked

Unless set by the user in terraform file, the default size can be hard to find:

- GCP :
  - if an image data resource exists, if user-provided credentials, Carbonifer lets terraform get the size of the disk image
  - if no data resource has been declared, a warning is printed out and as default, we use the following values (coming from empirical testing)
    - boot disk : 10 Gb, HDD
    - persistent disk: 500 Gb, HDD

### GPU

Similarily to [CPU](#cpu), GPU energy consumption is calculated from the GPU type from min/max Watt described in [Carbon Footprint Calculator](https://www.cloudcarbonfootprint.org/docs/methodology/#graphic-processing-units-gpus), we use min/max watt from constant file [GPU Watt per GPU Type](../data/gpu_watt.csv) and apply same formula as [CPU](#cpu).

Average GPU Utilization is also read from:

- user's config file in `$HOME/.carbonifer/config.yml`), variable `avg_gpu_use`
- targeted folder config file in `$TERRAFORM_PROJECT/.carbonifer/config.yml`), variable `avg_gpu_use`
- The default is `0.5` (50%)

## Carbon Intensity

This is the Carbon Emissions per Power per Time, in gCO2eq/Wh.

There are many models to get this Carbon Intensity:

- Declared by Cloud Provider
  - [Google](https://cloud.google.com/sustainability/region-carbon) shares it.
- From local Grid operator
  - Yearly average
  - Realtime (a good example is provided by [Electricity Maps](https://www.electricitymap.org/map))

Google claims net carbon emissions of all their regions are Zero, basically they compensate Carbon emissions of electricity used by investing in carbon offset. We decided to disregard this claim as electricity still comes from local Grid and are generated by actual Power plants running on Fossil fuels or renewable energy. The less electricity an infrastructure uses, the less it needs to be offset.

Currently, Carbonifer focuses on yearly average Grid carbon intensity, and we are using the following sources:

- [Google - 2021](https://github.com/GoogleCloudPlatform/region-carbon-info/blob/c154d6917e054d33380bb97098b7de8c0196a9f0/data/yearly/2021.csv)