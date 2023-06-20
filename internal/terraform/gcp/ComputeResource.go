package gcp

import (
	"github.com/carboniferio/carbonifer/internal/providers/gcp"
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/carboniferio/carbonifer/internal/terraform/tfrefs"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func getComputeResourceSpecs(
	resource *gjson.Result,
	tfRefs *tfrefs.References, zoneParam *string) *resources.ComputeResourceSpecs {

	machine_type := resource.Get("values.machine_type").String()

	zone := GetZones(resource)[0]

	machineType := gcp.GetGCPMachineType(machine_type, zone)
	CPUType := resource.Get("values.cpu_platform").String()

	var disks []disk

	bootDisks := resource.Get("values.boot_disk.#.initialize_params").Array()
	for _, bootDiskBlock := range bootDisks {
		disks = append(disks, getDisk(resource.Get("address").String(), &bootDiskBlock, true, tfRefs))
	}

	diskList := resource.Get("values.disk").Array()
	for _, diskBlock := range diskList {
		disks = append(disks, getDisk(resource.Get("address").String(), &diskBlock, false, tfRefs))
	}

	scratchDisks := resource.Get("values.scratch_disk").Array()
	for range scratchDisks {
		disks = append(disks, disk{isSSD: true, sizeGb: 375})
	}

	hddSize := decimal.Zero
	ssdSize := decimal.Zero
	for _, disk := range disks {
		if disk.isSSD {
			ssdSize = ssdSize.Add(decimal.NewFromFloat(disk.sizeGb))
		} else {
			hddSize = hddSize.Add(decimal.NewFromFloat(disk.sizeGb))
		}
	}

	gpus := machineType.GPUTypes

	gas := resource.Get("values.guest_accelerator").Array()
	for _, ga := range gas {
		gpuCount := ga.Get("count").Int()
		gpuType := ga.Get("type").String()
		for i := int64(0); i < gpuCount; i++ {
			gpus = append(gpus, gpuType)
		}
	}

	return &resources.ComputeResourceSpecs{
		GpuTypes:          gpus,
		VCPUs:             machineType.Vcpus,
		MemoryMb:          machineType.MemoryMb,
		CPUType:           CPUType,
		SsdStorage:        ssdSize,
		HddStorage:        hddSize,
		ReplicationFactor: 1,
	}
}
