package gcp

import (
	"github.com/carboniferio/carbonifer/internal/providers/gcp"
	"github.com/carboniferio/carbonifer/internal/resources"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/shopspring/decimal"
)

func getComputeResourceSpecs(
	resource tfjson.ConfigResource,
	dataResources *map[string]resources.DataResource, groupZone interface{}) *resources.ComputeResourceSpecs {

	machine_type := GetConstFromConfig(&resource, "machine_type").(string)
	var zone string
	if groupZone != nil {
		zone = groupZone.(string)
	} else {
		zone = GetConstFromConfig(&resource, "zone").(string)
	}

	machineType := gcp.GetGCPMachineType(machine_type, zone)
	CPUType, ok := GetConstFromConfig(&resource, "cpu_platform").(string)
	if !ok {
		CPUType = ""
	}

	var disks []disk
	bdExpr, ok_bd := resource.Expressions["boot_disk"]
	if ok_bd {
		bootDisks := bdExpr.NestedBlocks
		for _, bootDiskBlock := range bootDisks {
			bootDisk := getBootDisk(resource.Address, bootDiskBlock, dataResources)
			disks = append(disks, bootDisk)
		}
	}

	diskExpr, ok_bd := resource.Expressions["disk"]
	if ok_bd {
		disksBlocks := diskExpr.NestedBlocks
		for _, diskBlock := range disksBlocks {

			bootDisk := getDisk(resource.Address, diskBlock, false, dataResources)
			disks = append(disks, bootDisk)
		}
	}

	sdExpr, ok_sd := resource.Expressions["scratch_disk"]
	if ok_sd {
		scratchDisks := sdExpr.NestedBlocks
		for range scratchDisks {
			// Each scratch disk is 375GB
			//  source: https://cloud.google.com/compute/docs/disks#localssds
			disks = append(disks, disk{isSSD: true, sizeGb: 375})
		}
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
	gasI := GetConstFromConfig(&resource, "guest_accelerator")
	if gasI != nil {
		guestAccelerators := gasI.([]interface{})
		for _, gaI := range guestAccelerators {
			ga := gaI.(map[string]interface{})
			gpuCount := ga["count"].(float64)
			gpuType := ga["type"].(string)
			for i := float64(0); i < gpuCount; i++ {
				gpus = append(gpus, gpuType)
			}
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
