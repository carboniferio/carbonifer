// Get the list of all instances types of AWS and write them to a json to stdout with their attributes (cpu, memory, etc).

package main

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// InstanceType is the struct that will be exported in the json
type InstanceType struct {
	InstanceType string
	VCPU         int64
	MemoryMb     int64
	GPUs         []string
	GPUMemoryMb  int64
}

// Generate writes the list of instances types in a json to stdout
func main() {
	// Create a EC2 service client.
	session, err := session.NewSession(&aws.Config{Region: aws.String("us-east-1")})
	if err != nil {
		panic(err)
	}
	svc := ec2.New(session)

	// Get the list of instance types
	// Convert the list of instance types to the InstanceType struct
	instances := map[string]InstanceType{}
	token := describeInstanceTypesPaginated(svc, &instances, nil)
	for token != nil {
		token = describeInstanceTypesPaginated(svc, &instances, token)
	}

	// Write the list of instances to stdout
	json, err := json.MarshalIndent(instances, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func describeInstanceTypesPaginated(svc *ec2.EC2, instances *map[string]InstanceType, token *string) *string {
	instanceTypesOutput, err := svc.DescribeInstanceTypes(&ec2.DescribeInstanceTypesInput{
		NextToken: token,
	})
	if err != nil {
		panic(err)
	}

	for _, instanceType := range instanceTypesOutput.InstanceTypes {
		gpuInfos := instanceType.GpuInfo
		totalGPUMemoryMb := int64(0)
		gpus := []string{}
		if gpuInfos != nil {
			for _, gpu := range gpuInfos.Gpus {
				gpus = append(gpus, *gpu.Name)
			}
			totalGPUMemoryMb = *gpuInfos.TotalGpuMemoryInMiB
		}
		name := *instanceType.InstanceType
		instance := InstanceType{
			InstanceType: name,
			VCPU:         *instanceType.VCpuInfo.DefaultVCpus,
			MemoryMb:     *instanceType.MemoryInfo.SizeInMiB,
			GPUs:         gpus,
			GPUMemoryMb:  int64(totalGPUMemoryMb),
		}
		instanceMap := *instances
		instanceMap[name] = instance
	}
	return instanceTypesOutput.NextToken
}
