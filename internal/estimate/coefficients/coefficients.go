package coefficients

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Coefficients struct {
	CPUMinWh       decimal.Decimal `json:"cpu_min_wh"`
	CPUMaxWh       decimal.Decimal `json:"cpu_max_wh"`
	StorageHddWhTb decimal.Decimal `json:"storage_hdd_wh_tb"`
	StorageSsdWhTb decimal.Decimal `json:"storage_ssd_wh_tb"`
	NetworkingWhGb decimal.Decimal `json:"networking_wh_gb"`
	MemoryWhGb     decimal.Decimal `json:"memory_wh_gb"`
	PueAverage     decimal.Decimal `json:"pue_average"`
}

type CoefficientsProviders struct {
	AWS   Coefficients `json:"AWS"`
	GCP   Coefficients `json:"GCP"`
	Azure Coefficients `json:"Azure"`
}

var coefficientsPerProviders *CoefficientsProviders

func GetEnergyCoefficients() *CoefficientsProviders {
	if coefficientsPerProviders == nil {
		energyCoefFile := filepath.Join(viper.GetString("data.path"), "energy_coefficients.json")
		log.Debugf("reading Energy Coefficient Data file from: %v", energyCoefFile)
		jsonFile, err := os.Open(energyCoefFile)
		if err != nil {
			log.Fatal(err)
		}
		defer jsonFile.Close()

		byteValue, _ := io.ReadAll(jsonFile)
		err = json.Unmarshal([]byte(byteValue), &coefficientsPerProviders)
		if err != nil {
			log.Fatal(err)
		}
	}
	return coefficientsPerProviders
}

func (cps *CoefficientsProviders) GetByName(name string) Coefficients {
	r := reflect.ValueOf(cps)
	coefficients := reflect.Indirect(r).FieldByName(name)
	return coefficients.Interface().(Coefficients)
}
