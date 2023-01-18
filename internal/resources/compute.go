package resources

import (
	"fmt"

	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/shopspring/decimal"
)

type ComputeResourceSpecs struct {
	Gpu               int32
	HddStorage        decimal.Decimal
	SsdStorage        decimal.Decimal
	MemoryMb          int32
	VCPUs             int32
	CPUType           string
	ReplicationFactor int32
}

type ComputeResourceIdentification struct {
	// Indentification
	Name         string
	ResourceType string
	Provider     providers.Provider
	Region       string
}

type ComputeResource struct {
	Identification *ComputeResourceIdentification
	Specs          *ComputeResourceSpecs
}

func (r ComputeResource) IsSupported() bool {
	return true
}

func (r ComputeResource) GetIndentification() *ComputeResourceIdentification {
	return r.Identification
}

func (r ComputeResource) GetAddress() string {
	return fmt.Sprintf("%v.%v", r.GetIndentification().ResourceType, r.GetIndentification().Name)
}

type UnsupportedResource struct {
	Identification *ComputeResourceIdentification
}

func (r UnsupportedResource) IsSupported() bool {
	return false
}

func (r UnsupportedResource) GetIndentification() *ComputeResourceIdentification {
	return r.Identification
}

func (r UnsupportedResource) GetAddress() string {
	return fmt.Sprintf("%v.%v", r.GetIndentification().ResourceType, r.GetIndentification().Name)
}

type Resource interface {
	IsSupported() bool
	GetIndentification() *ComputeResourceIdentification
	GetAddress() string
}
