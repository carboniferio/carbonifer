package terraform

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/polkeli/yaml/v3" // TODO use go-yaml https://github.com/go-yaml/yaml/issues/100#issuecomment-1632853107
)

// Mapping is the mapping of the terraform resources
var globalMappings *Mappings

// GetMapping returns the mapping of the terraform resources
func getMapping(provider providers.Provider) (*Mappings, error) {
	if globalMappings != nil {
		return globalMappings, nil
	}
	err := loadMapping(provider)
	if err != nil {
		return nil, err
	}
	return globalMappings, nil
}

func loadMapping(provider providers.Provider) error {
	if globalMappings != nil {
		return nil
	}
	mappingFolder := fmt.Sprintf("internal/terraform/%s/mappings", provider)
	files, err := os.ReadDir(mappingFolder)
	if err != nil {
		return err
	}

	mergedMappings := &Mappings{
		General:         &GeneralConfig{},
		ComputeResource: &map[string]ResourceMapping{},
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		yamlFile, err := os.ReadFile(filepath.Join(mappingFolder, file.Name()))
		if err != nil {
			return err
		}
		var currentMapping Mappings
		err = yaml.Unmarshal(yamlFile, &currentMapping)
		if err != nil {
			return err
		}

		if currentMapping.General != nil {
			mergedMappings.General = currentMapping.General
		}

		if currentMapping.ComputeResource != nil {
			for k, v := range *currentMapping.ComputeResource {
				(*mergedMappings.ComputeResource)[k] = v
			}
		}

	}

	globalMappings = mergedMappings

	return nil
}
