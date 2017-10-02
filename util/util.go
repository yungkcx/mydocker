package util

import (
	"fmt"
	"os"
)

// Some directory names.
const (
	MydockerRootDir = "/var/lib/mydocker/"
	ImagesDir       = MydockerRootDir + "images/"
	ContainersDir   = MydockerRootDir + "containers/"
)

// PathExist returns true if path is exist, or false else.
func PathExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("Error judging whether %s exists: %v", path, err)
}
