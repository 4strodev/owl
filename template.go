package octopus

import "os"

type Template struct {
	Path string
	Content []os.FileInfo
}

