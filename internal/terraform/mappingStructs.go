package terraform

type ComputeResource struct {
	Resources map[string]Resource `yaml:",inline"`
}

type Resource struct {
	Paths      []string                      `yaml:"paths"`
	Type       string                        `yaml:"type"`
	Variables  map[string]VariableConfig     `yaml:"variables,omitempty"`
	Properties map[string]PropertyDefinition `yaml:"properties"`
}

type VariableConfig struct {
	Paths     []string  `yaml:"paths"`
	Reference Reference `yaml:"reference"`
}

type PropertyDefinition struct {
	Paths     interface{} `yaml:"paths"`
	Unit      string      `yaml:"unit,omitempty"`
	Default   interface{} `yaml:"default,omitempty"`
	Type      string      `yaml:"type,omitempty"`
	Reference *Reference  `yaml:"reference,omitempty"`
	Regex     *Regex      `yaml:"regex,omitempty"`
	Item      *Item       `yaml:"item,omitempty"`
}

type Reference struct {
	General    string   `yaml:"general,omitempty"`
	JsonFile   string   `yaml:"json_file,omitempty"`
	Property   string   `yaml:"property,omitempty"`
	Paths      []string `yaml:"paths,omitempty"`
	ReturnPath bool     `yaml:"return_path,omitempty"`
}

type Regex struct {
	Pattern string `yaml:"pattern"`
	Group   int    `yaml:"group"`
	Type    string `yaml:"type,omitempty"`
}

type Item struct {
	Count ItemDetail `yaml:"count"`
	Type  ItemDetail `yaml:"type"`
}

type ItemDetail struct {
	Paths string `yaml:"paths"`
	Type  string `yaml:"type"`
}
