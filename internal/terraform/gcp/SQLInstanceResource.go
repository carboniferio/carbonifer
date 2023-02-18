package gcp

import (
	"github.com/carboniferio/carbonifer/internal/providers/gcp"
	"github.com/carboniferio/carbonifer/internal/resources"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

func getSQLResourceSpecs(
	resource tfjson.ConfigResource) *resources.ComputeResourceSpecs {

	replicationFactor := int32(1)
	ssdSize := decimal.Zero
	hddSize := decimal.Zero
	var tier gcp.SqlTier

	settingsExpr, ok := resource.Expressions["settings"]
	if ok {
		settings := settingsExpr.NestedBlocks[0]

		availabilityType := settings["availability_type"]
		if availabilityType.ConstantValue != nil && availabilityType.ConstantValue == "REGIONAL" {
			replicationFactor = int32(2)
		}

		tierName := ""
		if settings["tier"] != nil {
			tierName = settings["tier"].ConstantValue.(string)
		}
		tier = gcp.GetGCPSQLTier(tierName)

		diskTypeI, ok_dt := settings["disk_type"]
		diskType := "PD_SSD"
		if ok_dt {
			diskType = diskTypeI.ConstantValue.(string)
		}

		diskSizeI, ok_ds := settings["disk_size"]
		diskSize := decimal.NewFromFloat(10)
		if ok_ds {
			diskSize = decimal.NewFromFloat(diskSizeI.ConstantValue.(float64))
		}

		if diskType == "PD_SSD" {
			ssdSize = diskSize
		} else if diskType == "PD_HDD" {
			hddSize = diskSize
		} else {
			log.Fatalf("%s : wrong type of disk : %s", resource.Address, tierName)
		}

	}

	return &resources.ComputeResourceSpecs{
		VCPUs:             int32(tier.Vcpus),
		MemoryMb:          int32(tier.MemoryMb),
		SsdStorage:        ssdSize,
		HddStorage:        hddSize,
		ReplicationFactor: replicationFactor,
	}
}
