package aws

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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
		instancesDataFile := filepath.Join(viper.GetString("data.path"), "aws_instances.json")
		log.Debugf("  reading aws instances data from: %v", instancesDataFile)
		jsonFile, err := os.Open(instancesDataFile)
		if err != nil {
			log.Fatal(err)
		}
		defer jsonFile.Close()

		byteValue, _ := io.ReadAll(jsonFile)
		err = json.Unmarshal([]byte(byteValue), &awsInstanceTypes)
		if err != nil {
			log.Fatal(err)
		}
	}

	return awsInstanceTypes[instanceTypeStr]
}
