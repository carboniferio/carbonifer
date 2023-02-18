package gcp

import (
	"github.com/carboniferio/carbonifer/internal/estimate/coefficients"
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/shopspring/decimal"
)

func estimateWattMem(resource *resources.ComputeResource) decimal.Decimal {
	return decimal.NewFromInt32(resource.Specs.MemoryMb).Div(decimal.NewFromInt32(1024)).Mul(coefficients.GetEnergyCoefficients().GCP.MemoryWhGb)
}
