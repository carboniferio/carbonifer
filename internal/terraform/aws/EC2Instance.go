package aws

import (
	"strings"

	"github.com/carboniferio/carbonifer/internal/providers/aws"
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/carboniferio/carbonifer/internal/terraform/tfrefs"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func getEC2Instance(
	resource tfjson.StateResource,
	tfRefs *tfrefs.References, groupZone interface{}) *resources.ComputeResourceSpecs {

	instanceType := resource.AttributeValues["instance_type"].(string)

	awsInstanceType := aws.GetAWSInstanceType(instanceType)

	var disks []disk

	amiId := ""
	if resource.AttributeValues["ami"] != nil {
		amiId = resource.AttributeValues["ami"].(string)
	}

	imageI := tfRefs.DataResources[amiId]

	var image *resources.AmiDataResource
	if imageI != nil {
		i := imageI.(resources.AmiDataResource)
		image = &i
	}

	// Root block device
	bd, ok_rd := resource.AttributeValues["root_block_device"]
	if ok_rd {
		rootDevices := bd.([]interface{})
		for _, rootDevice := range rootDevices {
			rootDisk := getDisk(resource.Address, rootDevice.(map[string]interface{}), true, image)
			disks = append(disks, rootDisk)
		}
	} else {
		if image != nil {
			rootDisk := disk{
				sizeGb: int64(image.DataImageSpecs[0].DiskSizeGb),
				isSSD:  IsSSD(image.DataImageSpecs[0].VolumeType),
			}
			disks = append(disks, rootDisk)
		} else {
			// Default root device
			rootDisk := disk{
				sizeGb: viper.GetInt64("provider.aws.disk.size"),
				isSSD:  true,
			}
			log.Warnf("No root device found for %s, using default root device of %vgb", resource.Address, viper.GetInt64("provider.aws.disk.size"))
			disks = append(disks, rootDisk)
		}
	}

	// Elastic block devices
	bd, ok_ebd := resource.AttributeValues["ebs_block_device"]
	if ok_ebd {
		ebds := bd.([]interface{})
		for _, blockDevice := range ebds {
			blockDisk := getDisk(resource.Address, blockDevice.(map[string]interface{}), false, image)
			disks = append(disks, blockDisk)
		}
	}

	// Ephemeral block devices
	epbd, ok_epbd := resource.AttributeValues["ephemeral_block_device"]
	if ok_epbd {
		epbds := epbd.([]interface{})
		for range epbds {
			instanceStorage := disk{
				sizeGb: int64(awsInstanceType.InstanceStorage.SizePerDiskGB),
				isSSD:  strings.ToLower(awsInstanceType.InstanceStorage.Type) == "ssd",
			}
			disks = append(disks, instanceStorage)
		}
	}

	hddSize := decimal.Zero
	ssdSize := decimal.Zero
	for _, disk := range disks {
		if disk.isSSD {
			ssdSize = ssdSize.Add(decimal.NewFromInt(disk.sizeGb))
		} else {
			hddSize = hddSize.Add(decimal.NewFromInt(disk.sizeGb))
		}
	}

	return &resources.ComputeResourceSpecs{
		VCPUs:             awsInstanceType.VCPU,
		MemoryMb:          awsInstanceType.MemoryMb,
		SsdStorage:        ssdSize,
		HddStorage:        hddSize,
		ReplicationFactor: 1,
	}
}
