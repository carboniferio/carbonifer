package gcp

import (
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/carboniferio/carbonifer/internal/terraform/tfrefs"
	"github.com/carboniferio/carbonifer/internal/utils"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/tidwall/gjson"
)

func getComputeDiskResourceSpecs(
	resource *gjson.Result,
	tfRefs *tfrefs.References) *resources.ComputeResourceSpecs {

	diskBlock := resource.Get("values")
	disk := getDisk(resource.Get("address").String(), &diskBlock, false, tfRefs)
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

func getDisk(resourceAddress string, diskBlock *gjson.Result, isBootDiskParam bool, tfRefs *tfrefs.References) disk {
	disk := disk{
		sizeGb:            0,
		isSSD:             true,
		replicationFactor: 1,
	}

	// Is Boot disk
	isBootDisk := isBootDiskParam
	if diskBlock.Get("boot").Bool() {
		isBootDisk = true
	}

	// Get disk type
	diskType := utils.GetOr(diskBlock, []string{"disk_type", "type"}).String()
	if diskType == "" {
		if isBootDisk {
			diskType = viper.GetString("provider.gcp.boot_disk.type")
		} else {
			diskType = viper.GetString("provider.gcp.disk.type")
		}
	}

	if diskType == "pd-standard" {
		disk.isSSD = false
	}

	// Get Disk size
	imageRes := diskBlock.Get("image")
	if imageRes.Exists() {
		image := (tfRefs.DataResources)[diskBlock.Get("image").String()]
		disk.sizeGb = (image.(resources.DataImageResource)).DataImageSpecs[0].DiskSizeGb
	}
	declaredSizeRes := utils.GetOr(diskBlock, []string{"size", "disk_size_gb"})
	if declaredSizeRes.Exists() {
		disk.sizeGb = declaredSizeRes.Float()
	}
	if disk.sizeGb == 0 {
		if isBootDisk {
			disk.sizeGb = viper.GetFloat64("provider.gcp.boot_disk.size")
		} else {
			disk.sizeGb = viper.GetFloat64("provider.gcp.disk.size")
		}
		log.Warningf("%v : Disk does not have a size declared, considering it default to be %vGb ", resourceAddress, disk.sizeGb)
	}

	// Get replication factor
	disk.replicationFactor = int32(len(GetZones(diskBlock)))

	return disk
}
