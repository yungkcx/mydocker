package main

import (
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/yungkcx/mydocker/cgroups"
	"github.com/yungkcx/mydocker/cgroups/subsystems"
	"github.com/yungkcx/mydocker/container"
)

// Run runs command with console if tty.
func Run(tty bool, comArray []string, res *subsystems.ResourceConfig) {
	parent, writePipe := container.NewParentProcess(tty)
	if parent == nil {
		log.Errorf("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Errorf("Error parent starting: %v", err)
	}
	cgroupManager := cgroups.NewCgroupManager("mydocker-cgroup")
	defer cgroupManager.Destroy()
	cgroupManager.Set(res)
	cgroupManager.Apply(parent.Process.Pid)

	log.Infoln("Parent process started, sending commands")
	sendInitCommand(comArray, writePipe)
	parent.Wait()
}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Infof("Command all is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}
