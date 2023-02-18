package resources

import (
	"fmt"

	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/shopspring/decimal"
)

type ComputeResourceSpecs struct {
	GpuTypes          []string
	HddStorage        decimal.Decimal
	SsdStorage        decimal.Decimal
	MemoryMb          int32
	VCPUs             int32
	CPUType           string
	ReplicationFactor int32
}

type ResourceIdentification struct {
	// Indentification
	Name         string
	ResourceType string
	Provider     providers.Provider
	Region       string
	SelfLink     string
	Count        int64
}

type ComputeResource struct {
	Identification *ResourceIdentification
	Specs          *ComputeResourceSpecs
}

func (r ComputeResource) IsSupported() bool {
	return true
}

func (r ComputeResource) GetIdentification() *ResourceIdentification {
	return r.Identification
}

func (r ComputeResource) GetAddress() string {
	return fmt.Sprintf("%v.%v", r.GetIdentification().ResourceType, r.GetIdentification().Name)
}

type UnsupportedResource struct {
	Identification *ResourceIdentification
}

func (r UnsupportedResource) IsSupported() bool {
	return false
}

func (r UnsupportedResource) GetIdentification() *ResourceIdentification {
	return r.Identification
}

func (r UnsupportedResource) GetAddress() string {
	return fmt.Sprintf("%v.%v", r.GetIdentification().ResourceType, r.GetIdentification().Name)
}

type Resource interface {
	IsSupported() bool
	GetIdentification() *ResourceIdentification
	GetAddress() string
}
