package resources

import (
	"fmt"

	internalProvider "github.com/carboniferio/carbonifer/internal/providers"
	"github.com/carboniferio/carbonifer/internal/providers/gcp"
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/carboniferio/carbonifer/internal/utils"
	"github.com/carboniferio/carbonifer/pkg/providers"
	"github.com/shopspring/decimal"
)

type GenericResource struct {
	Name              string
	Region            string
	Provider          providers.Provider
	GPUTypes          []string
	CPUTypes          []string
	VCPUs             int32
	MemoryMb          int32
	Storage           Storage
	ReplicationFactor int32
}

// IsSupported returns true if the resource is supported by carbonifer. At the moment, only GCP is supported.
func (g GenericResource) IsSupported() bool {
	// Use a switch to make it easier to add new providers
	switch g.Provider {
	case providers.GCP:
		return true
	default:
		return false
	}
}

func (g GenericResource) GetIdentification() *resources.ResourceIdentification {
	return &resources.ResourceIdentification{
		Name:         g.Name,
		ResourceType: "compute",
		Provider:     internalProvider.Provider(g.Provider),
		Region:       g.Region,
		Count:        1,
	}
}

func (g GenericResource) GetAddress() string {
	return fmt.Sprintf("%v.%v", g.GetIdentification().ResourceType, g.GetIdentification().Name)
}

type Storage struct {
	HddStorage decimal.Decimal
	SsdStorage decimal.Decimal
}

func GetResource(instanceType string, zone string, provider providers.Provider) (GenericResource, error) {
	switch provider {
	case providers.GCP:
		return fromGCPMachineTypeToResource(zone, gcp.GetGCPMachineType(instanceType, zone)), nil
	default:
		return GenericResource{}, fmt.Errorf("provider %s not supported", provider.String())
	}
}

func fromGCPMachineTypeToResource(region string, machineType gcp.MachineType) GenericResource {
	return GenericResource{
		Name:              machineType.Name,
		Region:            region,
		Provider:          providers.GCP,
		GPUTypes:          machineType.GPUTypes,
		MemoryMb:          machineType.MemoryMb,
		CPUTypes:          machineType.CpuTypes,
		VCPUs:             machineType.Vcpus,
		Storage:           Storage{},
		ReplicationFactor: 0,
	}
}

func init() {
	utils.InitWithDefaultConfig()
}
