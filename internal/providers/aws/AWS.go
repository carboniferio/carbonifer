package aws

import (
	"encoding/json"

	"github.com/carboniferio/carbonifer/internal/data"
	log "github.com/sirupsen/logrus"
)

type MachineType struct {
	InstanceType string `json:"InstanceType"`
	VCPU         int32  `json:"VCPU"`
	MemoryMb     int32  `json:"MemoryMb"`
}

var awsInstanceTypes map[string]MachineType

func GetAWSInstanceType(instanceTypeStr string) MachineType {
	log.Debugf("  Getting info for AWS machine type: %v", instanceTypeStr)
	if awsInstanceTypes == nil {
		byteValue := data.ReadDataFile("aws_instances.json")
		err := json.Unmarshal([]byte(byteValue), &awsInstanceTypes)
		if err != nil {
			log.Fatal(err)
		}
	}

	return awsInstanceTypes[instanceTypeStr]
}
