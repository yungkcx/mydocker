package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/yungkcx/mydocker/images"

	log "github.com/Sirupsen/logrus"
	"github.com/yungkcx/mydocker/cgroups"
	"github.com/yungkcx/mydocker/cgroups/subsystems"
	"github.com/yungkcx/mydocker/container"
	"github.com/yungkcx/mydocker/util"
)

// Run runs command with console if tty.
func Run(image string, tty bool, volume string, comArray []string, res *subsystems.ResourceConfig) error {
	// Create container's directory and lower-id file.
	if exist, err := images.IsImageExist(image); !exist {
		if err != nil {
			return err
		}
		return fmt.Errorf("Image %s is no exist", image)
	}
	rootURL := filepath.Join(util.ContainersDir, image)
	if err := os.MkdirAll(rootURL, 0700); err != nil {
		return fmt.Errorf("Error: %v", err)
	}
	parent, writePipe := container.NewParentProcess(tty, volume, rootURL)
	if err := ioutil.WriteFile(filepath.Join(rootURL, "lower-id"), []byte(image), 0744); err != nil {
		return fmt.Errorf("Error write to %s: %v", filepath.Join(rootURL, "lower-id"), err)
	}

	if parent == nil {
		return fmt.Errorf("New parent process error")
	}
	if err := parent.Start(); err != nil {
		return fmt.Errorf("Error parent starting: %v", err)
	}
	cgroupManager := cgroups.NewCgroupManager("mydocker-cgroup")
	defer cgroupManager.Destroy()
	cgroupManager.Set(res)
	cgroupManager.Apply(parent.Process.Pid)
	defer container.DeleteWorkSpace(rootURL)

	log.Infoln("Parent process started, sending commands")
	comArray = append([]string{volume, rootURL}, comArray...)
	sendInitCommand(comArray, writePipe)
	parent.Wait()

	return nil
}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Infof("Command all is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}
