package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	log "github.com/sirupsen/logrus"

	"google.golang.org/api/sqladmin/v1"

	"github.com/carboniferio/carbonifer/internal/providers/gcp"
	tools_gcp "github.com/carboniferio/carbonifer/internal/tools/gcp"
)

func getVCPUs(tierName string) (int64, error) {
	tierRegex := regexp.MustCompile(`db-(?P<class>[[:alpha:]]+\d+)-(?P<type>\w+)(-(?P<vcpus>\d+))?`)
	if tierRegex.MatchString(tierName) {
		values := tierRegex.FindAllStringSubmatch(tierName, -1)[0]
		if values[4] == "" {
			return 1, nil
		} else {
			vCPUs, err := strconv.Atoi(values[4])
			if err != nil {
				log.Fatalf(err.Error())
			}
			return int64(vCPUs), nil
		}
	} else {
		m := fmt.Sprintf("Cannot find number of vCPUs from tier name: %s", tierName)
		return 0, errors.New(m)
	}
}

func main() {

	ctx := context.Background()

	// Create a new Compute Engine client
	client, err := sqladmin.NewService(ctx)
	if err != nil {
		log.Fatalf("Error creating Cloud SQL client: %v", err)
	}

	project := tools_gcp.GetProjectId()

	tiersList := make(map[string]gcp.SqlTier)
	tiers, err := client.Tiers.List(project).Do()
	if err != nil {
		log.Fatal(err)
	}
	for _, tier := range tiers.Items {
		vCpus, err := getVCPUs(tier.Tier)
		if err != nil {
			log.Fatal(err)
		}
		tiersList[tier.Tier] = gcp.SqlTier{
			Name:        tier.Tier,
			Vcpus:       vCpus,
			MemoryMb:    tier.RAM / 1024 / 1024,
			DiskQuotaGB: tier.DiskQuota / 1024 / 1024 / 1024,
		}
	}

	// Generate the JSON representation of the list
	jsonData, err := json.MarshalIndent(tiersList, "", "  ")
	if err != nil {
		log.Fatalf("Error generating JSON: %v", err)
	}

	fmt.Println(string(jsonData))
}
