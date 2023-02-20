package gcp

import (
	"github.com/carboniferio/carbonifer/internal/estimate/coefficients"
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/shopspring/decimal"
)

func estimateWattStorage(resource *resources.ComputeResource) decimal.Decimal {
	provider := resource.Identification.Provider.String()
	storageSsdWhGb := coefficients.GetEnergyCoefficients().GetByName(provider).StorageSsdWhTb.Div(decimal.NewFromInt32(1024))
	storageHddWhGb := coefficients.GetEnergyCoefficients().GetByName(provider).StorageHddWhTb.Div(decimal.NewFromInt32(1024))
	storageSSDWh := resource.Specs.SsdStorage.Mul(storageSsdWhGb)
	storageHddWh := resource.Specs.HddStorage.Mul(storageHddWhGb)
	return storageSSDWh.Add(storageHddWh)
}
