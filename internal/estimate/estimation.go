package estimate

import (
	"time"

	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/shopspring/decimal"
)

type EstimationReport struct {
	Info      EstimationInfo
	Resources []EstimationResource
	Total     EstimationTotal
}

type EstimationResource struct {
	Resource        resources.ComputeResource
	Power           decimal.Decimal
	CarbonEmissions decimal.Decimal
	AverageCPUUsage decimal.Decimal
}

type EstimationTotal struct {
	Power           decimal.Decimal
	CarbonEmissions decimal.Decimal
	ResourcesCount  int
}

type EstimationInfo struct {
	UnitTime                string
	UnitWattTime            string
	UnitCarbonEmissionsTime string
	DateTime                time.Time
}
