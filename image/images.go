package image

import "github.com/yungkcx/mydocker/util"

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

func getImages() ([]string, error) {
	if exist, err := util.PathExist(util.ImagesDir); !exist {
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
	f, err := os.Open(util.ImagesDir)
	if err != nil {
		return nil, fmt.Errorf("Error: %v", err)
	}
	names, err := f.Readdirnames(-1)
	if err != nil {
		return nil, fmt.Errorf("Error read ImagesDir: %v", err)
	}
	return names, nil
}

// IsImageExist returns true if image is exist, or false else.
func IsImageExist(imageName string) (bool, error) {
	names, err := getImages()
	if err != nil {
		return false, err
	}

	for _, name := range names {
		if imageName == name {
			return true, nil
		}
	}
	return false, nil
}

// ListImages list images.
func ListImages() error {
	names, err := getImages()
	if err != nil {
		return err
	}
	// Print image directories.
	for _, name := range names {
		fmt.Println(name)
	}

	return nil
}

// CreateImage create an image directory named `name`` from given tar file.
func CreateImage(tar string, name string) error {
	if exist, err := util.PathExist(tar); !exist {
		if err != nil {
			return err
		}
		return fmt.Errorf("The tar file is not exist: %s", tar)
	}

	// Execute tar command to ImagesDir
	if name == "" {
		name = filepath.Base(strings.TrimSuffix(tar, path.Ext(tar)))
	}
	imageDir := filepath.Join(util.ImagesDir, name)
	if err := os.MkdirAll(imageDir, 0700); err != nil {
		return fmt.Errorf("Error %v", err)
	}
	cmd := exec.Command("tar", "xf", tar, "-C", imageDir)
	cmd.Stdout = ioutil.Discard
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Error running %s: %v", tar, err)
	}

	return nil
}

// RemoveImage removes an image.
func RemoveImage(names ...string) error {
	for _, name := range names {
		dir := filepath.Join(util.ImagesDir, name)
		if exist, err := util.PathExist(dir); !exist {
			if err != nil {
				return err
			}
			continue
		}
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("Error remove image %s: %v", dir, err)
		}
		fmt.Println(name)
	}
	return nil
}
