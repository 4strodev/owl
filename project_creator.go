package owl

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/4strodev/owl/template"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

const (
	// Permission MODES
	DIR_MODE int = 0755

	// FS errors
	DIR_EXISTS          string = "Directory exists"
	TEMPLATE_NOT_PASSED string = "No template passed"
	TEMPLATE_NOT_FOUND  string = "No template found"
)

// Contains the project's config
type ProjectConfig struct {
	Name               string
	fullPath           string
	TemplateName       string
	LocalTemplatesDirs []string
	VerboseOutput      bool
}

// Contains the config the fs api and template info
// fs: It manage all fs operations copy files and directories, read files, etc.
// Config: It saves project config: project name, templates directories, template name...
// template: saves the template struct that manage template operations: read template content,
// load and run template scripts.
type Project struct {
	fs       afero.Fs
	Config   ProjectConfig
	template template.ProjectTemplate
}

// Generates a new projects giving the project config and and a template config
func NewProject(projectConfig ProjectConfig, templateConfig template.TemplateConfig) *Project {
	return &Project{
		fs:     afero.NewOsFs(),
		Config: projectConfig,
		template: template.ProjectTemplate{
			Config: templateConfig,
		},
	}
}

// Create the instantiated project
func (self *Project) Create() error {
	// Getting folder name
	var err error

	// Loading template information and scripts if exist
	err = self.loadTemplate()
	if err != nil {
		return err
	}

	// Getting working directory for scripts
	WorkingDirectory, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// Getting full path for the directory of the project
	self.Config.fullPath = path.Join(WorkingDirectory, self.Config.Name)

	// Creating application folder
	err = self.CreateRootFolder(self.Config.fullPath)
	if err != nil {
		return err
	}

	if self.Config.VerboseOutput {
		fmt.Printf("Running on create scripts")
	}
	os.Chdir(self.Config.fullPath)
	self.template.RunOnCreateScripts()

	// Copying template content
	self.copyDir(self.template.Config.Path, self.Config.fullPath)

	if self.Config.VerboseOutput {
		fmt.Printf("Running on mount scripts")
	}
	os.Chdir(self.Config.fullPath)
	self.template.RunOnMountScripts()

	fmt.Printf("Project created succesfully\n")
	return nil
}

// Check if template is found and loads its content
func (self *Project) loadTemplate() error {
	var err error
	if self.Config.VerboseOutput {
		fmt.Printf("Loading template\n")
	}
	if self.Config.TemplateName == "" {
		err = fmt.Errorf(TEMPLATE_NOT_PASSED)
	}

	if self.Config.VerboseOutput {
		fmt.Printf("Searching template %s locally\n", self.Config.TemplateName)
	}
	// Searching template locally and loading data
	err = self.searchLocalTemplate(self.Config.LocalTemplatesDirs, self.Config.TemplateName)
	if err != nil {
		return fmt.Errorf(TEMPLATE_NOT_FOUND)
	}

	if self.Config.VerboseOutput {
		fmt.Printf("Loading scripts\n")
	}
	// Loading commands from config file
	err = self.template.LoadScripts()

	return err
}

// Search template in a local folder and load it's content
func (self *Project) searchLocalTemplate(directories []string, templateName string) error {
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

			// checking if folder has the same name as template
			if fileInfo.Name() == self.Config.TemplateName {
				// Reading folder content
				self.template.Content, err = afero.ReadDir(self.fs, path.Join(dir, fileInfo.Name()))
				if err != nil {
					log.Panicf("Cannot open templates dir: %s\n", err)
				}
				// saving template path
				self.template.Config.Path = path.Join(dir, fileInfo.Name())
				templateFound = true

				// setting config to viper using template config fields
				self.template.Viper = viper.New()
				self.template.Viper.AddConfigPath(self.template.Config.Path)
				self.template.Viper.SetConfigName(self.template.Config.ConfigName)
				self.template.Viper.SetConfigType(self.template.Config.ConfigType)
			}
		}
	}

	if templateFound {
		return nil
	}

	return fmt.Errorf(TEMPLATE_NOT_FOUND)
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

	if self.Config.VerboseOutput {
		fmt.Printf("Creating project folder\n")
	}
	err = self.fs.Mkdir(path, os.FileMode(DIR_MODE))
	if err != nil {
		log.Panicf("Cannot create directory project due this error: %s", err)
	}

	return nil
}

// Copy a target directory to a destination directory recursively
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
		// Copy each subdirectory
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
