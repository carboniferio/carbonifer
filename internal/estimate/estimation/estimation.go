package estimation

import (
	"time"

	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/shopspring/decimal"
)

// EstimationReport is the struct that contains the estimation report
type EstimationReport struct {
	Info                 EstimationInfo
	Resources            []EstimationResource
	UnsupportedResources []resources.Resource
	Total                EstimationTotal
}

// EstimationResource is the struct that contains the estimation of a resource
type EstimationResource struct {
	Resource        resources.Resource
	Power           decimal.Decimal `json:"PowerPerInstance"`
	CarbonEmissions decimal.Decimal `json:"CarbonEmissionsPerInstance"`
	AverageCPUUsage decimal.Decimal
	TotalCount      decimal.Decimal `json:"TotalCount"` // Count * ReplicationFactor
}

// EstimationTotal is the struct that contains the total estimation
type EstimationTotal struct {
	Power           decimal.Decimal
	CarbonEmissions decimal.Decimal
	ResourcesCount  decimal.Decimal
}

// EstimationInfo is the struct that contains the info of the estimation
type EstimationInfo struct {
	UnitTime                string
	UnitWattTime            string
	UnitCarbonEmissionsTime string
	DateTime                time.Time
	InfoByProvider          map[providers.Provider]InfoByProvider
}

// InfoByProvider is the struct that contains the info of the estimation by provider
type InfoByProvider struct {
	AverageCPUUsage float64
	AverageGPUUsage float64
}
