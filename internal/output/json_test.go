package output

import (
	"io/ioutil"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/carboniferio/carbonifer/internal/estimate"
	"github.com/carboniferio/carbonifer/internal/testutils"
	_ "github.com/carboniferio/carbonifer/internal/testutils"
)

func TestGenerateReportJson_Empty(t *testing.T) {
	wd := path.Join(testutils.RootDir, "test/terraform/nothing")
	viper.Set("workdir", wd)
	now := time.Now()

	estimations := estimate.EstimationReport{
		Info: estimate.EstimationInfo{
			UnitTime:                "h",
			UnitWattTime:            "w",
			UnitCarbonEmissionsTime: "gCO2eq/h",
			DateTime:                now,
		},
		Resources: []estimate.EstimationResource{},
		Total: estimate.EstimationTotal{
			Power:           decimal.Decimal{},
			CarbonEmissions: decimal.Decimal{},
			ResourcesCount:  0,
		},
	}

	want := loadOutput("nothing.txt")
	got := GenerateReportText(estimations)

	assert.Equal(t, strings.TrimSpace(want), strings.TrimSpace(got))
}

func loadOutput(name string) string {
	content, err := ioutil.ReadFile(path.Join(testutils.RootDir, "test/outputs", name))
	if err != nil {
		log.Fatal(err)
	}

	return string(content)

}
