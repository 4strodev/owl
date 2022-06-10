package owl

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/spf13/afero"
)

const (
	// Permission MODES
	DIR_MODE int = 0755

	// FS errors
	DIR_EXISTS          string = "Directory exists"
	TEMPLATE_NOT_PASSED string = "No template passed"
	TEMPLATE_NOT_FOUND  string = "No template found"
)

type ProjectConfig struct {
	Name string
	TemplateName string
	LocalTemplatesDirs []string
}

type Project struct {
	fs       afero.Fs
	Config   ProjectConfig
	template projectTemplate
}

func NewProject(projectConfig ProjectConfig, templateConfig TemplateConfig) *Project {
	return &Project{
		fs:     afero.NewOsFs(),
		Config: projectConfig,
		template: projectTemplate{
			config: templateConfig,
		},
	}
}

// Create the instantiated project
func (self *Project) Create() error {
	// Getting folder name
	var err error

	// Loading template information if exist
	fmt.Printf("Searching template '%s'\n", self.Config.TemplateName)
	err = self.loadTemplate()
	if err != nil {
		return err
	}

	// Creating application folder
	fmt.Printf("Creating folder '%s'\n", self.Config.Name)
	err = self.CreateRootFolder(self.Config.Name)
	if err != nil {
		return err
	}

	// TODO execute on create commands
	fmt.Printf("Executing on create commands\n")

	// Copying template content
	fmt.Printf("Copying template\n")
	self.copyDir(self.template.path, self.Config.Name)

	// TODO execute on mount commands
	fmt.Printf("Executing on mount commands\n")

	fmt.Printf("Project craeted succesfully\n")

	return nil
}

func (self *Project) loadTemplate() error {
	var err error
	if self.Config.TemplateName == "" {
		err = fmt.Errorf(TEMPLATE_NOT_PASSED)
	}

	// Searching template locally and loading data
	self.template, err = self.searchLocalTemplate(self.Config.LocalTemplatesDirs, self.Config.TemplateName)

	self.getCommandsFromConfigFile()

	// Loading commands from config file
	self.template.loadCommands()

	if err != nil {
		err = fmt.Errorf(TEMPLATE_NOT_FOUND)
	}

	return err
}

// Search template in a local folder and return it
func (self *Project) searchLocalTemplate(directories []string, templateName string) (projectTemplate, error) {
	var template projectTemplate
	var templateFound bool

	for _, dir := range directories {
		// Reading templates directory
		fileInfoList, err := afero.ReadDir(self.fs, dir)
		if err != nil {
			log.Panicf("Cannot open templates dir: %s\n", err)
		}

		for _, fileInfo := range fileInfoList {
			if !fileInfo.IsDir() {
				continue
			}

			if fileInfo.Name() == "owl_config.toml" {
				// TODO load template config file
			}

			// checking if folder has the same name as template
			if fileInfo.Name() == self.Config.TemplateName {
				// Reading folder content
				template.content, err = afero.ReadDir(self.fs, path.Join(dir, fileInfo.Name()))
				if err != nil {
					log.Panicf("Cannot open templates dir: %s\n", err)
				}
				// saving template path
				template.path = path.Join(dir, fileInfo.Name())
				templateFound = true

				// setting config to viper using template config fields
				template.viper.AddConfigPath(template.path)
				template.viper.SetConfigName(template.config.ConfigName)
				template.viper.SetConfigType(template.config.ConfigType)
			}
		}
	}

	if templateFound {
		return template, nil
	}

	return projectTemplate{}, fmt.Errorf(TEMPLATE_NOT_FOUND)
}

// Search template in a repo and return it
func (self *Project) searchRemoteTemplate(template string) bool {
	return false
}

// Create the root folder of the project
func (self *Project) CreateRootFolder(path string) error {
	// Create project folder if not exists
	exists, err := afero.DirExists(self.fs, path)
	if err != nil {
		panic(err)
	}

	if exists {
		return fmt.Errorf(DIR_EXISTS)
	}

	err = self.fs.Mkdir(path, os.FileMode(DIR_MODE))
	if err != nil {
		log.Panicf("Cannot create directory project due this error: %s", err)
	}

	return nil
}

func (self *Project) copyDir(targetDirPath string, destination string) {
	var err error
	var pendingDirs []os.FileInfo

	// Reading directory content
	targetContent, err := afero.ReadDir(self.fs, targetDirPath)
	if err != nil {
		log.Panicf("Error reading %s: %s", targetDirPath, err)
	}

	for _, fileInfo := range targetContent {
		// Path is the directory dirPath that contains the files
		if fileInfo.IsDir() {
			// Adding dir to pending directories
			pendingDirs = append(pendingDirs, fileInfo)
			continue
		}

		// Copying only files for the current directory
		destinationFilePath := path.Join(destination, fileInfo.Name())
		// Copying file using the same permissions
		destinationFile, err := self.fs.OpenFile(destinationFilePath, os.O_RDWR|os.O_CREATE, fileInfo.Mode())
		if err != nil {
			log.Panicf("Error while creating %s file: %s\n", destinationFilePath, err.Error())
		}

		// Reading the target file and copying content to destination file
		targetfilePath := path.Join(targetDirPath, fileInfo.Name())
		targetFileContent, err := afero.ReadFile(self.fs, targetfilePath)
		_, err = destinationFile.Write(targetFileContent)
		if err != nil {
			log.Panicf("Error while writing to file %s: %s", destinationFile.Name(), err)
		}

		// Closing now to avoid memory overflow
		destinationFile.Close()
	}

	// If there are pending directories
	if len(pendingDirs) > 0 {
		// Copy each directory
		for _, dir := range pendingDirs {
			targetDirPath := path.Join(targetDirPath, dir.Name())
			destinationDirPath := path.Join(destination, dir.Name())

			err = self.fs.Mkdir(destinationDirPath, os.FileMode(dir.Mode()))
			if err != nil {
				log.Panicf("Error craeting dir %s: %s", destinationDirPath, err)
			}

			self.copyDir(targetDirPath, destinationDirPath)
		}
	}
}

func (self *Project) getCommandsFromConfigFile() ([2]templateCommands, error) {
	return [2]templateCommands{
		{},
		{},
	}, nil
}
