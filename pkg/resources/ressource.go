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

// GenericResource is a struct that contains the information of a generic resource
type GenericResource struct {
	Address           string
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

// GetIdentification returns the identification of the resource
func (g GenericResource) GetIdentification() *resources.ResourceIdentification {
	return &resources.ResourceIdentification{
		Name:         g.Name,
		ResourceType: "compute",
		Provider:     internalProvider.Provider(g.Provider),
		Region:       g.Region,
		Count:        1,
		Address:      g.Address,
	}
}

// GetAddress returns the address of the resource
func (g GenericResource) GetAddress() string {
	return g.Address
}

// Storage is the struct that contains the storage of a resource
type Storage struct {
	HddStorage decimal.Decimal
	SsdStorage decimal.Decimal
}

// GetResource returns a GenericResource from an instance type
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
		CPUTypes:          machineType.CPUTypes,
		VCPUs:             machineType.Vcpus,
		Storage:           Storage{},
		ReplicationFactor: 0,
	}
}

func init() {
	utils.InitWithDefaultConfig()
}
