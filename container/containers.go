package container

import (
	"fmt"
	"os"

	"github.com/yungkcx/mydocker/util"
)

func getContainers() ([]string, error) {
	if exist, err := util.PathExist(util.ContainersDir); !exist {
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
	f, err := os.Open(util.ContainersDir)
	if err != nil {
		return nil, fmt.Errorf("Error: %v", err)
	}
	names, err := f.Readdirnames(-1)
	if err != nil {
		return nil, fmt.Errorf("Error read ContainersDir: %v", err)
	}
	return names, nil
}

// ListContainers list containers.
func ListContainers() error {
	names, err := getContainers()
	if err != nil {
		return err
	}
	// Print image directories.
	for _, name := range names {
		fmt.Println(name)
	}

	return nil
}
