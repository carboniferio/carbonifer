package terraform

import (
	"os"
	"path/filepath"

	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/polkeli/yaml/v3" // TODO use go-yaml https://github.com/go-yaml/yaml/issues/100#issuecomment-1632853107
)

// Mapping is the mapping of the terraform resources
var globalMappings *Mappings

// GetMapping returns the mapping of the terraform resources
func getMapping() (*Mappings, error) {
	if globalMappings != nil {
		return globalMappings, nil
	}
	err := loadMappings()
	if err != nil {
		return nil, err
	}
	return globalMappings, nil
}

func loadMappings() error {
	mappingsPath := "internal/terraform/mappings"
	files, err := os.ReadDir(mappingsPath)
	if err != nil {
		return err
	}

	// Iterate over each entry
	for _, file := range files {
		// Check if it's a directory
		if file.IsDir() {
			// Get the relative path
			relativePath := filepath.Join(mappingsPath, file.Name())

			// Process the subfolder
			err := loadMapping(relativePath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func loadMapping(providerMappingFolder string) error {
	files, err := os.ReadDir(providerMappingFolder)
	if err != nil {
		return err
	}

	mergedMappings := &Mappings{
		General:         &map[providers.Provider]GeneralConfig{},
		ComputeResource: &map[string]ResourceMapping{},
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		yamlFile, err := os.ReadFile(filepath.Join(providerMappingFolder, file.Name()))
		if err != nil {
			return err
		}
		var currentMapping Mappings
		err = yaml.Unmarshal(yamlFile, &currentMapping)
		if err != nil {
			return err
		}

		if currentMapping.General != nil {
			for k, v := range *currentMapping.General {
				(*mergedMappings.General)[k] = v
			}
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
