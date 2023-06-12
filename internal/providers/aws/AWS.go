package aws

import (
	"encoding/json"

	"github.com/carboniferio/carbonifer/internal/data"
	log "github.com/sirupsen/logrus"
)

type InstanceType struct {
	InstanceType    string          `json:"InstanceType"`
	VCPU            int32           `json:"VCPU"`
	MemoryMb        int32           `json:"MemoryMb"`
	InstanceStorage InstanceStorage `json:"InstanceStorage"`
}

type InstanceStorage struct {
	SizePerDiskGB int64 `json:"SizePerDiskGB"`
	Count         int32 `json:"Count"`
	Type          string
}

var awsInstanceTypes map[string]InstanceType

func GetAWSInstanceType(instanceTypeStr string) InstanceType {
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
