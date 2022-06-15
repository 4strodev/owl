package template

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/spf13/viper"
)

// The full template struct
type ProjectTemplate struct {
	Content         []os.FileInfo
	Viper           *viper.Viper
	Config          TemplateConfig
	onCreateScripts TemplateScripts
	onMountScripts  TemplateScripts
}

// Read the config file and load the data into template config
func (self *ProjectTemplate) LoadScripts() error {
	var err error
	err = self.Viper.ReadInConfig()
	onCreateScripts := map[string]scriptsMap{
		"universal": {},
		"linux":     {},
		"windows":   {},
		"macos":     {},
	}

	onMountScripts := map[string]scriptsMap{
		"universal": {},
		"linux":     {},
		"windows":   {},
		"macos":     {},
	}

	// TODO this code is bullshit
	for key := range onCreateScripts {
		onCreateScripts[key], err = self.parseScripts(
			self.Viper.GetStringMapStringSlice(fmt.Sprintf("scripts.oncreate.%s",key)),
			self.Config.Context,
		)
		if err != nil {
			return fmt.Errorf("Error parsing %s: %s", key, err)
		}
	}

	for key := range onMountScripts {
		onMountScripts[key], err = self.parseScripts(
			self.Viper.GetStringMapStringSlice(fmt.Sprintf("scripts.onmount.%s",key)),
			self.Config.Context,
		)
		if err != nil {
			return fmt.Errorf("Error parsing %s: %s", key, err)
		}
	}

	fmt.Printf("On create scripts: %v\n", onCreateScripts)
	fmt.Printf("On mount scripts: %v\n", onMountScripts)

	return nil
}

func (self *ProjectTemplate) parseScripts(rawScripts scriptsMap, context map[string]any) (map[string][]string, error) {
	parsedScripts := make(map[string][]string)

	for key, scriptSequence := range rawScripts {
		var parsedScriptSequence []string

		for _, command := range scriptSequence {
			// Creating variables
			var templateBuffer bytes.Buffer
			template := template.New(key)

			// Parsing template and return possible errors
			_, err := template.Parse(command)
			if err != nil {
				return nil, err
			}

			// Executing template and saving result in a buffer
			template.Execute(&templateBuffer, context)

			parsedScriptSequence = append(parsedScriptSequence, string(templateBuffer.Bytes()))
		}

		// Changing raw script by parsed script
		parsedScripts[key] = parsedScriptSequence
	}
	return parsedScripts, nil
}

// Execute on create scripts
func (self *ProjectTemplate) RunOnCreateScripts() error {
	var err error
	// TODO execute loaded commands
	return err
}

// Execute on mount scripts
func (self *ProjectTemplate) RunOnMountScripts() error {
	var err error
	// TODO execute loaded commands
	return err
}