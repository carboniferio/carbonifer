package gcp

import (
	"github.com/carboniferio/carbonifer/internal/resources"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func getComputeDiskResourceSpecs(
	resource tfjson.ConfigResource,
	dataResources *map[string]resources.DataResource) *resources.ComputeResourceSpecs {

	disk := getDisk(resource.Address, resource.Expressions, false, dataResources)
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

func getBootDisk(resourceAddress string, bootDiskBlock map[string]*tfjson.Expression, dataResources *map[string]resources.DataResource) disk {
	var disk disk
	initParams := bootDiskBlock["initialize_params"].NestedBlocks
	for _, initParam := range initParams {
		disk = getDisk(resourceAddress, initParam, true, dataResources)

	}
	return disk
}

func getDisk(resourceAddress string, diskBlock map[string]*tfjson.Expression, isBootDiskParam bool, dataResources *map[string]resources.DataResource) disk {
	disk := disk{
		sizeGb:            viper.GetFloat64("provider.gcp.boot_disk.size"),
		isSSD:             true,
		replicationFactor: 1,
	}

	// Is Boot disk
	isBootDisk := isBootDiskParam
	isBootDiskI := GetConstFromExpression(diskBlock["boot"])
	if isBootDiskI != nil {
		isBootDisk = isBootDiskI.(bool)
	}

	// Get disk type
	var diskType string
	diskTypeExpr := diskBlock["type"]
	if diskTypeExpr == nil {
		if isBootDisk {
			diskType = viper.GetString("provider.gcp.boot_disk.type")
		} else {
			diskType = viper.GetString("provider.gcp.disk.type")
		}
	} else {
		diskType = diskTypeExpr.ConstantValue.(string)
	}

	if diskType == "pd-standard" {
		disk.isSSD = false
	}

	// Get Disk size
	declaredSize := GetConstFromExpression(diskBlock["size"])
	if declaredSize == nil {
		declaredSize = GetConstFromExpression(diskBlock["disk_size_gb"])
	}
	if declaredSize == nil {
		if isBootDisk {
			disk.sizeGb = viper.GetFloat64("provider.gcp.boot_disk.size")
		} else {
			disk.sizeGb = viper.GetFloat64("provider.gcp.disk.size")
		}
		diskImageLinkExpr, okImage := diskBlock["image"]
		if okImage {
			for _, ref := range diskImageLinkExpr.References {
				image, ok := (*dataResources)[ref]
				if ok {
					disk.sizeGb = (image.(resources.DataImageResource)).DataImageSpecs.DiskSizeGb
				} else {
					log.Warningf("%v : Disk image does not have a size declared, considering it default to be 10Gb ", resourceAddress)
				}
			}
		} else {
			log.Warningf("%v : Boot disk size not declared. Please set it! (otherwise we assume 10gb) ", resourceAddress)

		}
	} else {
		disk.sizeGb = declaredSize.(float64)
	}

	replicaZonesExpr := diskBlock["replica_zones"]
	if replicaZonesExpr != nil {
		rz := replicaZonesExpr.ConstantValue.([]interface{})
		disk.replicationFactor = int32(len(rz))
	} else {
		disk.replicationFactor = 1
	}

	return disk
}
