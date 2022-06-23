package template

// template constants
const (
	TOML string = "toml"
	YAML string = "yaml"
	JSON string = "json"
)

// Set config for your configuration
type TemplateConfig struct {
	Path       string
	ConfigType string
	ConfigName string
	Context    map[string]any
}
