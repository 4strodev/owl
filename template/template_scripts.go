package template

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Alias to make easier to read
type scriptsMap map[string][]string

// A set of scripts to execute
type TemplateScripts struct {
	UniversalScripts map[string][]exec.Cmd // `scripts_scope:"universal"`
	LinuxScripts     map[string][]exec.Cmd // `scripts_scope:"linux"`
	MacosScripts     map[string][]exec.Cmd // `scripts_scope:"macos"`
	WindowsScripts   map[string][]exec.Cmd // `scripts_scope:"windows"`
}

func (self *TemplateScripts) Unmarshal(scriptsScopeMap map[string]scriptsMap) error {
	var err error

	for key, scriptMap := range scriptsScopeMap {
		switch key {
		case "universal":
			self.UniversalScripts = parseMapScript(scriptMap)
			break
		case "linux":
			self.LinuxScripts = parseMapScript(scriptMap)
			break
		case "macos":
			self.MacosScripts = parseMapScript(scriptMap)
			break
		case "windows":
			self.WindowsScripts = parseMapScript(scriptMap)
			break
		}
	}
	return err
}

func (self *TemplateScripts) run() error {
	var err error
	for key, script := range self.UniversalScripts {
		err = runScript(script)
		if err != nil {
			return fmt.Errorf("Error executing %s script: %s", key, err)
		}
	}
	switch runtime.GOOS {
	case "linux":
		for key, script := range self.LinuxScripts {
			err = runScript(script)
			if err != nil {
				return fmt.Errorf("Error executing %s script: %s", key, err)
			}
		}
	case "windows":
		for key, script := range self.WindowsScripts {
			err = runScript(script)
			if err != nil {
				return fmt.Errorf("Error executing %s script: %s", key, err)
			}
		}
	case "darwin":
		for key, script := range self.MacosScripts {
			err = runScript(script)
			if err != nil {
				return fmt.Errorf("Error executing %s script: %s", key, err)
			}
		}
	}
	return err
}

func runScript(script []exec.Cmd) error {
	var err error
	for _, command := range script {
		command.Stdout = os.Stdout
		err = command.Run()
		if err != nil {
			return err
		}
	}
	return err
}

func parseMapScript(scriptMap scriptsMap) map[string][]exec.Cmd {
	parsedScriptsMap := make(map[string][]exec.Cmd)

	for mapKey, commandList := range scriptMap {
		parsedCommandList := parseCommandList(commandList)
		parsedScriptsMap[mapKey] = parsedCommandList
	}

	return parsedScriptsMap
}

func parseCommandList(commandList []string) []exec.Cmd {
	var parsedCommandList []exec.Cmd

	for _, command := range commandList {
		splitedCommand := strings.Split(command, " ")

		newCommand := exec.Command(splitedCommand[0], splitedCommand[1:]...)

		parsedCommandList = append(parsedCommandList, *newCommand)
	}

	return parsedCommandList
}
