// Get the list of all instances types of AWS and write them to a json to stdout with their attributes (cpu, memory, etc).

package main

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"
)

// InstanceType is the struct that will be exported in the json
type instanceType struct {
	InstanceType    string
	VCPU            int64
	MemoryMb        int64
	GPUs            []string
	GPUMemoryMb     int64
	InstanceStorage *instanceStorage
}

type instanceStorage struct {
	SizePerDiskGB int64
	Count         int64
	Type          string
}

// Generate writes the list of instances types in a json to stdout
func main() {
	// Create a EC2 service client.
	session, err := session.NewSession(&aws.Config{Region: aws.String("us-east-1")})
	if err != nil {
		errW := errors.Wrap(err, "cannot create aws session")
		log.Panic(errW)
	}
	svc := ec2.New(session)

	// Get the list of instance types
	// Convert the list of instance types to the InstanceType struct
	instances := map[string]instanceType{}
	token := describeInstanceTypesPaginated(svc, &instances, nil)
	for token != nil {
		token = describeInstanceTypesPaginated(svc, &instances, token)
	}

	// Write the list of instances to stdout
	json, err := json.MarshalIndent(instances, "", "  ")
	if err != nil {
		errW := errors.Wrap(err, "cannot marshal instances to json")
		log.Panic(errW)
	}
	fmt.Println(string(json))
}

func describeInstanceTypesPaginated(svc *ec2.EC2, instances *map[string]instanceType, token *string) *string {
	instanceTypesOutput, err := svc.DescribeInstanceTypes(&ec2.DescribeInstanceTypesInput{
		NextToken: token,
	})
	if err != nil {
		errW := errors.Wrap(err, "cannot describe instance types")
		log.Panic(errW)
	}

	for _, instanceTypeInfo := range instanceTypesOutput.InstanceTypes {
		gpuInfos := instanceTypeInfo.GpuInfo
		totalGPUMemoryMb := int64(0)
		gpus := []string{}
		if gpuInfos != nil {
			for _, gpu := range gpuInfos.Gpus {
				gpus = append(gpus, *gpu.Name)
			}
			totalGPUMemoryMb = *gpuInfos.TotalGpuMemoryInMiB
		}
		var instanceStorageInfo instanceStorage
		if instanceTypeInfo.InstanceStorageSupported != nil && *instanceTypeInfo.InstanceStorageSupported {
			instanceStorageInfo = instanceStorage{
				SizePerDiskGB: *instanceTypeInfo.InstanceStorageInfo.Disks[0].SizeInGB,
				Count:         *instanceTypeInfo.InstanceStorageInfo.Disks[0].Count,
				Type:          *instanceTypeInfo.InstanceStorageInfo.Disks[0].Type,
			}
		}
		name := *instanceTypeInfo.InstanceType
		instance := instanceType{
			InstanceType:    name,
			VCPU:            *instanceTypeInfo.VCpuInfo.DefaultVCpus,
			MemoryMb:        *instanceTypeInfo.MemoryInfo.SizeInMiB,
			GPUs:            gpus,
			GPUMemoryMb:     int64(totalGPUMemoryMb),
			InstanceStorage: &instanceStorageInfo,
		}
		instanceMap := *instances
		instanceMap[name] = instance
	}
	return instanceTypesOutput.NextToken
}
