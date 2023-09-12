package resources

import (
	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/shopspring/decimal"
)

// ComputeResourceSpecs is the struct that contains the specs of a compute resource
type ComputeResourceSpecs struct {
	GpuTypes   []string
	HddStorage decimal.Decimal
	SsdStorage decimal.Decimal
	MemoryMb   int32
	VCPUs      int32
	CPUType    string
}

// ResourceIdentification is the struct that contains the identification of a resource
type ResourceIdentification struct {
	// Indentification
	Name              string
	ResourceType      string
	Provider          providers.Provider
	Region            string
	Count             int64
	ReplicationFactor int32
	Address           string
}

// ComputeResource is the struct that contains the info of a compute resource
type ComputeResource struct {
	Identification *ResourceIdentification
	Specs          *ComputeResourceSpecs
}

// IsSupported returns true if the resource is supported, false otherwise
func (r ComputeResource) IsSupported() bool {
	return true
}

// GetIdentification returns the identification of the resource
func (r ComputeResource) GetIdentification() *ResourceIdentification {
	return r.Identification
}

// GetAddress returns the address of the resource
func (r ComputeResource) GetAddress() string {
	return r.Identification.Address
}

// UnsupportedResource is the struct that contains the info of an unsupported resource
type UnsupportedResource struct {
	Identification *ResourceIdentification
}

// IsSupported returns true if the resource is supported, false otherwise
func (r UnsupportedResource) IsSupported() bool {
	return false
}

// GetIdentification returns the identification of the resource
func (r UnsupportedResource) GetIdentification() *ResourceIdentification {
	return r.Identification
}

// GetAddress returns the address of the resource
func (r UnsupportedResource) GetAddress() string {
	return r.Identification.Address
}

// Resource is the interface that contains the info of a resource
type Resource interface {
	IsSupported() bool
	GetIdentification() *ResourceIdentification
	GetAddress() string
}
