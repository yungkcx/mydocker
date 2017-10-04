package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/yungkcx/mydocker/cgroups"
	"github.com/yungkcx/mydocker/cgroups/subsystems"
	"github.com/yungkcx/mydocker/container"
)

// Run runs command with console if tty.
func Run(imageName string, tty bool, volume string, cmdArray []string, res *subsystems.ResourceConfig) error {
	containerRoot, err := container.NewContainer(imageName)
	if err != nil {
		return err
	}
	// New containerProcess process.
	containerProcess, writePipe := container.NewParentProcess(tty, volume, containerRoot)
	defer container.DeleteWorkSpace(containerRoot)

	if containerProcess == nil {
		return fmt.Errorf("New containerProcess process error")
	}
	if err := containerProcess.Start(); err != nil {
		return fmt.Errorf("Error containerProcess starting: %v", err)
	}
	containerPid := containerProcess.Process.Pid
	if err := container.SetConfigFile(imageName, containerRoot, cmdArray, containerPid); err != nil {
		return err
	}

	cgroupManager := cgroups.NewCgroupManager("mydocker-cgroup")
	defer cgroupManager.Destroy()
	cgroupManager.Set(res)
	cgroupManager.Apply(containerPid)

	log.Infoln("Container process started, sending commands")
	cmdArray = append([]string{volume, containerRoot}, cmdArray...)
	sendInitCommand(cmdArray, writePipe)
	// Ingore SIGINT to destroy container.
	signal.Ignore(os.Interrupt)
	if tty {
		containerProcess.Wait()
		if err := container.ExitInfo(containerRoot); err != nil {
			return err
		}
	}

	return nil
}

func sendInitCommand(cmdArray []string, writePipe *os.File) {
	command := strings.Join(cmdArray, " ")
	writePipe.WriteString(command)
	writePipe.Close()
}
