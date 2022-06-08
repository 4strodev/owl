package octopus

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/spf13/afero"
)

const (
	// Permission MODES
	DIR_MODE  int = 0755
	FILE_MODE int = 0755

	// FS errors
	DIR_EXISTS          string = "Directory exists"
	TEMPLATE_NOT_PASSED string = "No template passed"
	TEMPLATE_NOT_FOUND  string = "No template found"
)

type Project struct {
	fs       afero.Fs
	Config   ProjectConfig
	Template Template
}

func NewProject(config ProjectConfig) *Project {
	return &Project{
		fs:     afero.NewOsFs(),
		Config: config,
	}
}

// Create the instantiated project
func (self *Project) Create() error {
	// Getting folder name
	folderName := fmt.Sprintf("./%s", self.Config.Name)
	var err error

	err = self.LoadTemplate()
	if err != nil {
		return err
	}

	err = self.CreateRootFolder(folderName)
	if err != nil {
		return err
	}

	err = self.CopyTemplate()
	if err != nil {
		return err
	}

	return nil
}

func (self *Project) LoadTemplate() error {
	var err error
	if self.Config.TemplateName == "" {
		err = fmt.Errorf(TEMPLATE_NOT_PASSED)
	}

	self.Template, err = self.SearchLocalTemplate(self.Config.LocalTemplatesDirs, self.Config.TemplateName)

	if err != nil {
		err = fmt.Errorf(TEMPLATE_NOT_FOUND)
	}

	return err
}

// Search template in a local folder and return it
func (self *Project) SearchLocalTemplate(directories []string, templateName string) (Template, error) {
	var template Template
	var templateFound bool

	for _, dir := range directories {
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
				template.Content, err = afero.ReadDir(self.fs, path.Join(dir, fileInfo.Name()))
				if err != nil {
					log.Panicf("Cannot open templates dir: %s\n", err)
				}
				template.Path = path.Join(dir, fileInfo.Name())
				templateFound = true
			}
		}
	}

	if templateFound {
		return template, nil
	}
	return Template{}, fmt.Errorf(TEMPLATE_NOT_FOUND)
}

// Search template in a repo and return it
func (self *Project) SearchRemoteTemplate(template string) bool {
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

func (self *Project) CopyTemplate() error {
	for _, fileInfo := range self.Template.Content {
		// Path is the directory dirPath that contains the files
		dirPath := self.Template.Path
		if fileInfo.IsDir() {
			// TODO if is a directory copy the content of the directory
			continue
		}

		destinationFile, err := self.fs.Create(path.Join(self.Config.Name, fileInfo.Name()))
		if err != nil {
			log.Panicf("Error while copying the content of the template %s\n", err.Error())
		}

		targetFileContent, err := afero.ReadFile(self.fs, path.Join(dirPath, fileInfo.Name()))
		_, err = destinationFile.Write(targetFileContent)
		if err != nil {
			log.Panicf("Error while writing to file %s: %s", destinationFile.Name(), err)
		}
	}
	return nil
}
