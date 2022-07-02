package owl

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"

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

	CONFIG_FILE_NAME string = "owl_config"
)

// Contains the project's config
type ProjectConfig struct {
	Name               string
	fullPath           string
	TemplateName       string
	LocalTemplatesDirs []string
	VerboseOutput      bool
	TempDir            string
	ignoreGlobs        []string
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
	if projectConfig.TempDir == "" {
		projectConfig.TempDir = "/tmp/owl"
	}

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

	// If tempdir is not created create them
	if _, err = self.fs.Stat(self.Config.TempDir); os.IsNotExist(err) {
		err = self.fs.Mkdir(self.Config.TempDir, os.FileMode(DIR_MODE))
		if err != nil {
			log.Panic(err)
		}
	}

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
		fmt.Printf("Running on create scripts\n")
	}
	os.Chdir(self.Config.fullPath)
	self.template.RunOnCreateScripts()

	// Copying template content
	self.copyDir(self.template.Config.Path, self.Config.fullPath)

	if self.Config.VerboseOutput {
		fmt.Printf("Running on mount scripts\n")
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
	if err == nil {
		if self.Config.VerboseOutput {
			fmt.Printf("Loading scripts\n")
		}
		// Loading commands from config file
		err = self.template.LoadScripts()

		return err
	}

	err = self.searchRemoteTemplate(self.Config.TemplateName)
	if err != nil {
		return err
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

	// Searching template on temporary dir
	// TODO recursive search

	// Searching template in provided folders
	for _, dir := range directories {
		templateFound = self.searchTemplateOnDir(dir, templateName)
	}

	if templateFound {
		return nil
	}

	return fmt.Errorf(TEMPLATE_NOT_FOUND)
}

// Check if template was downloaded previously
// download the template if does not exist
func (self *Project) searchRemoteTemplate(templateRepo string) error {
	var err error
	var stderr *bytes.Buffer = new(bytes.Buffer)
	var templateFound bool

	// Removing https://
	regexTemplateParser := regexp.MustCompile("https?://")
	parsedTemplateRepo := regexTemplateParser.ReplaceAllString(templateRepo, "")
	templateName := path.Base(parsedTemplateRepo)

	// Getting template parent dir path and the destination of the cloned template
	templateParentPath := path.Join(self.Config.TempDir, path.Dir(parsedTemplateRepo))
	cloneDestination := path.Join(self.Config.TempDir, parsedTemplateRepo)

	// Searching template on tmp dir
	templateFound = self.searchTemplateOnDir(templateParentPath, templateName)
	if templateFound {
		return err
	}

	// Creating command to clone repo
	cloneTemplate := exec.Command("git", "clone", templateRepo, cloneDestination)
	cloneTemplate.Stdout = os.Stdout
	// Capturing errors in a buffer
	// Because the error returned by commands
	// is a status code and not error message
	cloneTemplate.Stderr = stderr

	// Executing command
	if self.Config.VerboseOutput {
		fmt.Printf("Cloning %s\n", templateRepo)
	}
	err = cloneTemplate.Run()

	// If an error ocurred return the captured error message
	if err != nil {
		return fmt.Errorf(stderr.String())
	}

	// Search again the template on tmp dir
	templateFound = self.searchTemplateOnDir(templateParentPath, templateName)
	if !templateFound {
		return fmt.Errorf(TEMPLATE_NOT_FOUND)
	}

	return err
}

// Search a template
// used in searchLocalTemplate and searchRemoteTemplate
func (self *Project) searchTemplateOnDir(dir string, templateName string) bool {
	var templateFound bool

	// Reading templates directory
	fileInfoList, err := afero.ReadDir(self.fs, dir)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		//log.Panic(err)
	}

	for _, fileInfo := range fileInfoList {
		if !fileInfo.IsDir() {
			continue
		}

		// checking if folder has the same name as template
		if fileInfo.Name() == templateName {
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
			self.template.Viper.SetConfigName(CONFIG_FILE_NAME)
			self.template.Viper.SetConfigType(self.template.Config.ConfigType)
		}
	}

	return templateFound
}

// Create the root folder of the project
func (self *Project) CreateRootFolder(path string) error {
	var err error

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
	var pendingDirs []os.FileInfo // pendingDirs it's going to save the subdirectories paths

	// Reading directory content
	targetContent, err := afero.ReadDir(self.fs, targetDirPath)
	if err != nil {
		log.Panicf("Error reading %s: %s", targetDirPath, err)
	}

	for _, fileInfo := range targetContent {
		// Don't copy the config file
		if fileInfo.Name() == ".owlignore" {
			//self.fs.Open(path.Join(targetDirPath, fileInfo.Name()))
			continue
		}

		if fileInfo.Name() == ".git" {
			continue
		}

		if fileInfo.Name() == "owl_config.toml" {
			continue
		}

		// If it's a dir add to pending dirs stack to copy their content
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
