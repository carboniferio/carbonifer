package aws

import (
	"strings"

	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/carboniferio/carbonifer/internal/terraform/tfrefs"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type disk struct {
	sizeGb            int64
	isSSD             bool
	replicationFactor int32
}

func getDisk(resourceAddress string, diskBlock map[string]interface{}, isBootDisk bool, image *resources.EbsDataResource, trRefs *tfrefs.References) disk {
	disk := disk{
		sizeGb:            viper.GetInt64("provider.aws.boot_disk.size"),
		isSSD:             true,
		replicationFactor: 1,
	}

	// Get disk type

	diskType := viper.GetString("provider.aws.disk.type")
	diskTypeI := diskBlock["volume_type"]
	if diskTypeI == nil {
		diskTypeI = diskBlock["type"]
	}
	if diskTypeI != nil {
		diskType = diskTypeI.(string)
	} else {
		if image != nil && diskBlock["device_name"] != nil {
			if strings.HasPrefix(image.Identification.ResourceType, "aws_ebs_snapshot") {
				disk.sizeGb = int64(image.DataImageSpecs[0].DiskSizeGb)
			}
			if strings.HasPrefix(image.Identification.ResourceType, "aws_ami") {
				for _, bd := range image.DataImageSpecs {
					if bd != nil {
						if strings.HasPrefix(bd.DeviceName, diskBlock["device_name"].(string)) {
							diskType = bd.VolumeType
						}
					}
				}
			}
		}
	}

	disk.isSSD = IsSSD(diskType)

	// Get Disk size
	declaredSize := diskBlock["volume_size"]
	if declaredSize == nil {
		declaredSize = diskBlock["size"]
	}
	if declaredSize == nil && diskBlock["snapshot_id"] != nil {
		snapshotId := diskBlock["snapshot_id"].(string)
		snapshot := getAwsImage(trRefs, snapshotId)
		declaredSize = snapshot.DataImageSpecs[0].DiskSizeGb
	}
	if declaredSize == nil {
		if image != nil {
			// Case of snapshot, no device name
			if strings.HasPrefix(image.Identification.ResourceType, "aws_ebs_snapshot") {
				disk.sizeGb = int64(image.DataImageSpecs[0].DiskSizeGb)
			}
			// Case of ami, we use device name, except for boot disk
			if strings.HasPrefix(image.Identification.ResourceType, "aws_ami") {
				searchedDeviceName := "/dev/sda1"
				if !isBootDisk {
					searchedDeviceName = diskBlock["device_name"].(string)
				}
				for _, bd := range image.DataImageSpecs {
					if bd != nil {
						if strings.HasPrefix(bd.DeviceName, searchedDeviceName) {
							disk.sizeGb = int64(bd.DiskSizeGb)
						}
					}
				}
			}
		} else {
			disk.sizeGb = viper.GetInt64("provider.aws.disk.size")
			log.Warningf("%v : Disk size not declared. Please set it! (otherwise we assume %vsgb) ", resourceAddress, disk.sizeGb)

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

func getEbsVolume(tfResource tfjson.StateResource, tfRefs *tfrefs.References) *resources.ComputeResourceSpecs {

	// Get image if it comes from a snapshot
	var image *resources.EbsDataResource
	if tfResource.AttributeValues["snapshot_id"] != nil {
		awsImageId := tfResource.AttributeValues["snapshot_id"].(string)
		image = getAwsImage(tfRefs, awsImageId)
	}

	// Get disk specifications
	disk := getDisk(tfResource.Address, tfResource.AttributeValues, false, image, tfRefs)
	hddSize := decimal.Zero
	ssdSize := decimal.Zero
	if disk.isSSD {
		ssdSize = decimal.NewFromInt(disk.sizeGb)
	} else {
		hddSize = decimal.NewFromInt(disk.sizeGb)
	}
	computeResourceSpecs := resources.ComputeResourceSpecs{
		SsdStorage: ssdSize,
		HddStorage: hddSize,
	}

	return &computeResourceSpecs
}
