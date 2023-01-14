package resources

import (
	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/shopspring/decimal"
)

type ComputeResource struct {
	// Indentification
	Name         string
	ResourceType string
	Provider     providers.Provider
	Region       string

	// Size
	Gpu        int32
	HddStorage decimal.Decimal
	SsdStorage decimal.Decimal
	MemoryMb   int32
	VCPUs      int32
	CPUType    string
}
