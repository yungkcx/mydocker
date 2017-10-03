package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/yungkcx/mydocker/cgroups"
	"github.com/yungkcx/mydocker/cgroups/subsystems"
	"github.com/yungkcx/mydocker/container"
)

// Run runs command with console if tty.
func Run(imageName string, tty bool, volume string, comArray []string, res *subsystems.ResourceConfig) error {
	containerRoot, err := container.NewContainer(imageName)
	if err != nil {
		return err
	}
	// New containerProcess process.
	containerProcess, writePipe := container.NewParentProcess(tty, volume, containerRoot)
	if err := ioutil.WriteFile(filepath.Join(containerRoot, "lower-id"), []byte(imageName), 0744); err != nil {
		return fmt.Errorf("Error write to %s: %v", filepath.Join(containerRoot, "lower-id"), err)
	}
	defer container.DeleteWorkSpace(containerRoot)

	if containerProcess == nil {
		return fmt.Errorf("New containerProcess process error")
	}
	if err := containerProcess.Start(); err != nil {
		return fmt.Errorf("Error containerProcess starting: %v", err)
	}
	cgroupManager := cgroups.NewCgroupManager("mydocker-cgroup")
	defer cgroupManager.Destroy()
	cgroupManager.Set(res)
	cgroupManager.Apply(containerProcess.Process.Pid)

	log.Infoln("Container process started, sending commands")
	comArray = append([]string{volume, containerRoot}, comArray...)
	log.Infoln(comArray)
	sendInitCommand(comArray, writePipe)
	// Ingore SIGINT to destroy container.
	signal.Ignore(os.Interrupt)
	containerProcess.Wait()

	return nil
}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	writePipe.WriteString(command)
	writePipe.Close()
}
