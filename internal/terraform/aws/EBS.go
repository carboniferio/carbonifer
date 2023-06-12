package aws

import (
	"strings"

	"github.com/carboniferio/carbonifer/internal/resources"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type disk struct {
	sizeGb            int64
	isSSD             bool
	replicationFactor int32
}

func getDisk(resourceAddress string, diskBlock map[string]interface{}, isBootDisk bool, image *resources.AmiDataResource) disk {
	disk := disk{
		sizeGb:            viper.GetInt64("provider.aws.boot_disk.size"),
		isSSD:             true,
		replicationFactor: 1,
	}

	// Get disk type

	diskType := viper.GetString("provider.aws.disk.type")
	diskTypeI := diskBlock["volume_type"]
	if diskTypeI != nil {
		diskType = diskTypeI.(string)
	} else {
		if image != nil && diskBlock["device_name"] != nil {
			for _, bd := range image.DataImageSpecs {
				if strings.HasPrefix(bd.DeviceName, diskBlock["device_name"].(string)) {
					diskType = bd.VolumeType
				}
			}
		}
	}

	disk.isSSD = IsSSD(diskType)

	// Get Disk size
	declaredSize := diskBlock["volume_size"]
	if declaredSize == nil {
		if image != nil {
			for _, bd := range image.DataImageSpecs {
				if strings.HasPrefix(bd.DeviceName, "/dev/sda") {
					disk.sizeGb = int64(bd.DiskSizeGb)
				}
			}
		} else {
			disk.sizeGb = viper.GetInt64("provider.aws.disk.size")
			log.Warningf("%v : Boot disk size not declared. Please set it! (otherwise we assume %vsgb) ", resourceAddress, disk.sizeGb)

		}
	} else {
		disk.sizeGb = int64(declaredSize.(float64))
	}
	return disk
}

func IsSSD(diskType string) bool {
	isSSD := false
	if strings.HasPrefix(diskType, "gp") || strings.HasPrefix(diskType, "io") {
		isSSD = true
	}
	return isSSD
}
