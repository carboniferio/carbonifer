package estimation

import (
	"time"

	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/shopspring/decimal"
)

type EstimationReport struct {
	Info                 EstimationInfo
	Resources            []EstimationResource
	UnsupportedResources []resources.Resource
	Total                EstimationTotal
}

type EstimationResource struct {
	Resource        resources.Resource
	Power           decimal.Decimal `json:"PowerPerInstance"`
	CarbonEmissions decimal.Decimal `json:"CarbonEmissionsPerInstance"`
	AverageCPUUsage decimal.Decimal
	Count           decimal.Decimal
}

type EstimationTotal struct {
	Power           decimal.Decimal
	CarbonEmissions decimal.Decimal
	ResourcesCount  decimal.Decimal
}

type EstimationInfo struct {
	UnitTime                string
	UnitWattTime            string
	UnitCarbonEmissionsTime string
	DateTime                time.Time
	AverageCPUUsage         float64
	AverageGPUUsage         float64
}
