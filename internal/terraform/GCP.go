package terraform

import (
	"fmt"
	"strings"

	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/carboniferio/carbonifer/internal/resources"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func GetResource(tfResource tfjson.StateResource, dataResources *map[string]resources.DataResource) resources.Resource {
	resourceId := getResourceIdentification(tfResource)
	if resourceId.ResourceType == "google_compute_instance" {
		specs := getComputeResourceSpecs(tfResource, dataResources)
		return resources.ComputeResource{
			Identification: resourceId,
			Specs:          specs,
		}
	}
	if resourceId.ResourceType == "google_compute_disk" ||
		resourceId.ResourceType == "google_compute_region_disk" {
		specs := getComputeDiskResourceSpecs(tfResource, dataResources)
		return resources.ComputeResource{
			Identification: resourceId,
			Specs:          specs,
		}
	}
	return resources.UnsupportedResource{
		Identification: resourceId,
	}
}

func getResourceIdentification(resource tfjson.StateResource) *resources.ResourceIdentification {
	region := resource.AttributeValues["region"]
	if region == nil {
		if resource.AttributeValues["zone"] != nil {
			zone := resource.AttributeValues["zone"].(string)
			region = strings.Join(strings.Split(zone, "-")[:2], "-")
		} else if resource.AttributeValues["replica_zones"] != nil {
			replica_zones := resource.AttributeValues["replica_zones"].([]interface{})
			// should be all in the same region
			region = strings.Join(strings.Split(replica_zones[0].(string), "-")[:2], "-")
		} else {
			region = ""
		}
	}
	selfLink := ""
	if resource.AttributeValues["self_link"] != nil {
		selfLink = resource.AttributeValues["self_link"].(string)
	}

	return &resources.ResourceIdentification{
		Name:         resource.Name,
		ResourceType: resource.Type,
		Provider:     providers.GCP,
		Region:       fmt.Sprint(region),
		SelfLink:     selfLink,
	}
}

func getComputeResourceSpecs(
	resource tfjson.StateResource,
	dataResources *map[string]resources.DataResource) *resources.ComputeResourceSpecs {

	machine_type := resource.AttributeValues["machine_type"].(string)
	zone := resource.AttributeValues["zone"].(string)
	machineType := providers.GetGCPMachineType(machine_type, zone)
	CPUType, ok := resource.AttributeValues["cpu_platform"].(string)
	if !ok {
		CPUType = ""
	}

	var disks []disk
	bootDisks := resource.AttributeValues["boot_disk"].([]interface{})
	for _, bootDiskBlock := range bootDisks {
		bootDisk := getBootDisk(resource.Address, bootDiskBlock.(map[string]interface{}), dataResources)
		disks = append(disks, bootDisk)
	}

	scratchDisks := resource.AttributeValues["scratch_disk"].([]interface{})
	for range scratchDisks {
		// Each scratch disk is 375GB
		//  source: https://cloud.google.com/compute/docs/disks#localssds
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
	return &resources.ComputeResourceSpecs{
		Gpu:        machineType.Gpus,
		VCPUs:      machineType.Vcpus,
		MemoryMb:   machineType.MemoryMb,
		CPUType:    CPUType,
		SsdStorage: ssdSize,
		HddStorage: hddSize,
	}
}

func getComputeDiskResourceSpecs(
	resource tfjson.StateResource,
	dataResources *map[string]resources.DataResource) *resources.ComputeResourceSpecs {

	disk := getDisk(resource.Address, resource.AttributeValues, false, dataResources)
	hddSize := decimal.Zero
	ssdSize := decimal.Zero
	if disk.isSSD {
		ssdSize = ssdSize.Add(decimal.NewFromFloat(disk.sizeGb))
	} else {
		hddSize = hddSize.Add(decimal.NewFromFloat(disk.sizeGb))
	}
	return &resources.ComputeResourceSpecs{
		SsdStorage:        ssdSize,
		HddStorage:        hddSize,
		ReplicationFactor: disk.replicationFactor,
	}
}

type disk struct {
	sizeGb            float64
	isSSD             bool
	replicationFactor int32
}

func getBootDisk(resourceAddress string, bootDiskBlock map[string]interface{}, dataResources *map[string]resources.DataResource) disk {
	var disk disk
	initParams := bootDiskBlock["initialize_params"]
	for _, iP := range initParams.([]interface{}) {
		initParam := iP.(map[string]interface{})
		disk = getDisk(resourceAddress, initParam, true, dataResources)

	}
	return disk
}

func getDisk(resourceAddress string, diskBlock map[string]interface{}, isBootDisk bool, dataResources *map[string]resources.DataResource) disk {
	disk := disk{
		sizeGb:            viper.GetFloat64("provider.gcp.boot_disk.size"),
		isSSD:             true,
		replicationFactor: 1,
	}
	diskType := diskBlock["type"]
	if diskType == nil {
		if isBootDisk {
			diskType = viper.GetString("provider.gcp.boot_disk.type")
		} else {
			diskType = viper.GetString("provider.gcp.disk.type")
		}
	}
	if diskType == "pd-standard" {
		disk.isSSD = false
	}

	sizeParam := diskBlock["size"]
	if sizeParam != nil {
		disk.sizeGb = sizeParam.(float64)
	} else {
		if isBootDisk {
			disk.sizeGb = viper.GetFloat64("provider.gcp.boot_disk.size")
		} else {
			disk.sizeGb = viper.GetFloat64("provider.gcp.disk.size")
		}
		diskImageLink, ok := diskBlock["image"]
		if ok {
			image, ok := (*dataResources)[diskImageLink.(string)]
			if ok {
				disk.sizeGb = (image.(resources.DataImageResource)).DataImageSpecs.DiskSizeGb
			} else {
				log.Warningf("%v : Disk image does not have a size declared, considering it default to be 10Gb ", resourceAddress)
			}
		} else {
			log.Warningf("%v : Boot disk size not declared. Please set it! (otherwise we assume 10gb) ", resourceAddress)

		}

	}

	replicaZones := diskBlock["replica_zones"]
	if replicaZones != nil {
		rz := replicaZones.([]interface{})
		disk.replicationFactor = int32(len(rz))
	}

	return disk
}
