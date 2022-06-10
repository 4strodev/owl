package owl

import (
	"os"
	"os/exec"

	"github.com/spf13/viper"
)

// template constants
const (
	TOML string = "toml"
	YAML string = "yaml"
	JSON string = "json"
)

// Alias to make easier to read
type commandList []exec.Cmd

// A set of commands to execute
type templateCommands struct {
	universalCommands map[string]commandList
	linuxCommands     map[string]commandList
	macosCommands     map[string]commandList
	windowsCommands   map[string]commandList
}

// Set config for your configuration
type TemplateConfig struct {
	ConfigType       string
	ConfigName       string
	onCreateCommands templateCommands
	onMountCommands  templateCommands
}

// The full template struct
type projectTemplate struct {
	path    string
	content []os.FileInfo
	viper   *viper.Viper
	config  TemplateConfig
}

// Read the config file and load the data into template config
func (self *projectTemplate) loadCommands() error {
	return nil
}

// Execute on create commands
func (self *projectTemplate) runOnCreateCommands() error {
	var err error
	return err
}

// Execute on mount commands
func (self *projectTemplate) runOnMountCommands() error {
	var err error
	return err
}
