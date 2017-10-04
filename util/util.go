package util

import (
	"fmt"
	"os"
)

// Some directory names.
var (
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

// BKDRHash is a string hash function.
func BKDRHash(s string) uint64 {
	seed := uint64(131)
	hash := uint64(0)
	for _, ch := range s {
		hash = hash*seed + uint64(ch)
	}
	return hash & 0x7fffffff
}
