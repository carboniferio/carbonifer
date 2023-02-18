package gcp

import (
	"github.com/carboniferio/carbonifer/internal/providers/gcp"
	"github.com/carboniferio/carbonifer/internal/resources"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/shopspring/decimal"
)

func getComputeResourceSpecs(
	resource tfjson.StateResource,
	dataResources *map[string]resources.DataResource, groupZone interface{}) *resources.ComputeResourceSpecs {

	machine_type := resource.AttributeValues["machine_type"].(string)
	var zone string
	if groupZone != nil {
		zone = groupZone.(string)
	} else {
		zone = resource.AttributeValues["zone"].(string)
	}

	machineType := gcp.GetGCPMachineType(machine_type, zone)
	CPUType, ok := resource.AttributeValues["cpu_platform"].(string)
	if !ok {
		CPUType = ""
	}

	var disks []disk
	bd, ok_bd := resource.AttributeValues["boot_disk"]
	if ok_bd {
		bootDisks := bd.([]interface{})
		for _, bootDiskBlock := range bootDisks {
			bootDisk := getBootDisk(resource.Address, bootDiskBlock.(map[string]interface{}), dataResources)
			disks = append(disks, bootDisk)
		}
	}

	// TODO Disks
	diskListI, ok_disks := resource.AttributeValues["disk"]
	if ok_disks {
		diskList := diskListI.([]interface{})
		for _, diskBlock := range diskList {
			disk := getDisk(resource.Address, diskBlock.(map[string]interface{}), false, dataResources)
			disks = append(disks, disk)
		}
	}

	sd, ok_sd := resource.AttributeValues["scratch_disk"]
	if ok_sd {
		scratchDisks := sd.([]interface{})
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
	gasI, ok := resource.AttributeValues["guest_accelerator"]
	if ok {
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
