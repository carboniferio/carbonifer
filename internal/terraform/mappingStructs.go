package terraform

import "github.com/carboniferio/carbonifer/internal/providers"

type Mappings struct {
	General         *map[providers.Provider]GeneralConfig `yaml:"general,omitempty"`
	ComputeResource *map[string]ResourceMapping           `yaml:"compute_resource,omitempty"`
}

type GeneralConfig struct {
	JSONData         *map[string]interface{} `yaml:"json_data,omitempty"`
	DiskTypes        *DiskTypes              `yaml:"disk_types,omitempty"`
	IgnoredResources *[]string               `yaml:"ignored_resources,omitempty"`
}

type DiskTypes struct {
	Default *DiskType             `yaml:"default,omitempty"`
	Types   *map[string]*DiskType `yaml:"types,omitempty"`
}

type ResourceMapping struct {
	Paths      []string                         `yaml:"paths"`
	Type       string                           `yaml:"type"`
	Variables  *ResourceMapping                 `yaml:"variables,omitempty"`
	Properties *map[string][]PropertyDefinition `yaml:"properties"`
}

type PropertyDefinition struct {
	Paths     []string           `yaml:"paths"`
	Unit      *string            `yaml:"unit,omitempty"`
	Default   interface{}        `yaml:"default,omitempty"`
	ValueType *string            `yaml:"value_type,omitempty"`
	Reference *Reference         `yaml:"reference,omitempty"`
	Regex     *Regex             `yaml:"regex,omitempty"`
	Item      *[]ResourceMapping `yaml:"item,omitempty"`
}

type Reference struct {
	General    string   `yaml:"general,omitempty"`
	JSONFile   string   `yaml:"json_file,omitempty"`
	Property   string   `yaml:"property,omitempty"`
	Paths      []string `yaml:"paths,omitempty"`
	ReturnPath bool     `yaml:"return_path,omitempty"`
}

type Regex struct {
	Pattern string `yaml:"pattern"`
	Group   int    `yaml:"group"`
	Type    string `yaml:"type,omitempty"`
}
