package gcp

import (
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/carboniferio/carbonifer/internal/terraform/tfrefs"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func getComputeDiskResourceSpecs(
	resource tfjson.StateResource,
	tfRefs *tfrefs.References) *resources.ComputeResourceSpecs {

	disk := getDisk(resource.Address, resource.AttributeValues, false, tfRefs)
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

func getBootDisk(resourceAddress string, bootDiskBlock map[string]interface{}, tfRefs *tfrefs.References) disk {
	var disk disk
	initParams := bootDiskBlock["initialize_params"]
	for _, iP := range initParams.([]interface{}) {
		initParam := iP.(map[string]interface{})
		disk = getDisk(resourceAddress, initParam, true, tfRefs)

	}
	return disk
}

func getDisk(resourceAddress string, diskBlock map[string]interface{}, isBootDiskParam bool, tfRefs *tfrefs.References) disk {
	disk := disk{
		sizeGb:            viper.GetFloat64("provider.gcp.boot_disk.size"),
		isSSD:             true,
		replicationFactor: 1,
	}

	// Is Boot disk
	isBootDisk := isBootDiskParam
	isBootDiskI := diskBlock["boot"]
	if isBootDiskI != nil {
		isBootDisk = isBootDiskI.(bool)
	}

	// Get disk type
	var diskType string
	diskTypeExpr := diskBlock["type"]
	if diskTypeExpr == nil {
		diskTypeExpr = diskBlock["disk_type"]
	}
	if diskTypeExpr == nil {
		if isBootDisk {
			diskType = viper.GetString("provider.gcp.boot_disk.type")
		} else {
			diskType = viper.GetString("provider.gcp.disk.type")
		}
	} else {
		diskType = diskTypeExpr.(string)
	}

	if diskType == "pd-standard" {
		disk.isSSD = false
	}

	// Get Disk size
	declaredSize := diskBlock["size"]
	if declaredSize == nil {
		declaredSize = diskBlock["disk_size_gb"]
	}
	if declaredSize == nil {
		if isBootDisk {
			disk.sizeGb = viper.GetFloat64("provider.gcp.boot_disk.size")
		} else {
			disk.sizeGb = viper.GetFloat64("provider.gcp.disk.size")
		}
		diskImageLink := diskBlock["image"]
		if diskImageLink != nil {
			image, ok := (tfRefs.DataResources)[diskImageLink.(string)]
			if ok {
				disk.sizeGb = (image.(resources.DataImageResource)).DataImageSpecs[0].DiskSizeGb
			} else {
				log.Warningf("%v : Disk image does not have a size declared, considering it default to be 10Gb ", resourceAddress)
			}
		} else {
			log.Warningf("%v : Boot disk size not declared. Please set it! (otherwise we assume 10gb) ", resourceAddress)

		}
	} else {
		disk.sizeGb = declaredSize.(float64)
	}

	replicaZones := diskBlock["replica_zones"]
	if replicaZones != nil {
		rz := replicaZones.([]interface{})
		disk.replicationFactor = int32(len(rz))
	}

	return disk
}
