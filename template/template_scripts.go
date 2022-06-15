package template

import (
	"os/exec"
)

// Alias to make easier to read
type scriptsMap map[string][]string

// A set of scripts to execute
type TemplateScripts struct {
	UniversalScripts map[string]exec.Cmd `scripts_scope:"universal"`
	LinuxScripts     map[string]exec.Cmd `scripts_scope:"linux"`
	MacosScripts     map[string]exec.Cmd `scripts_scope:"macos"`
	WindowsScripts   map[string]exec.Cmd `scripts_scope:"windows"`
}

func (self *TemplateScripts) Unmarshal(scripts scriptsMap) error {
	var err error
	return err
}


