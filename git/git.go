package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

func Clone(repo, destination string) error {
	var stderr *bytes.Buffer = new(bytes.Buffer)
	var err error

	// Creating command to clone repo
	cloneTemplate := exec.Command("git", "clone", repo, destination)
	cloneTemplate.Stdout = os.Stdout
	// Capturing errors in a buffer
	// Because the error returned by commands
	// is a status code and not error message
	cloneTemplate.Stderr = stderr

	err = cloneTemplate.Run()

	// If an error ocurred return the captured error message
	if err != nil {
		return fmt.Errorf(stderr.String())
	}

	return nil
}
