package container

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/yungkcx/mydocker/util"
)

// RunContainerInitProcess is a function.
func RunContainerInitProcess() error {
	args := readUserCommand()
	log.Infof("Command all is %s", args[2:])
	volume, rootURL := args[0], args[1]
	if err := setupMount(volume, rootURL); err != nil {
		return err
	}

	containerCommand := args[2]
	path, err := exec.LookPath(containerCommand)
	if err != nil {
		return fmt.Errorf("Error lookpath %s: %v", containerCommand, err)
	}
	if err := syscall.Exec(path, args[2:], nil); err != nil {
		return fmt.Errorf("Error exec container command: %v", err)
	}
	return nil
}

func readUserCommand() []string {
	pipe := os.NewFile(uintptr(3), "pipe")
	defer pipe.Close()
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		log.Errorf("init read pipe error %f", err)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}

func realChroot(path string) error {
	if err := syscall.Chroot(path); err != nil {
		return fmt.Errorf("error after fallback to chroot: %v", err)
	}
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("error changing to new root after chroot: %v", err)
	}
	return nil
}

func pivotRoot(root string) error {
	// Chdir to root.
	root = filepath.Join(root, "merged")
	if err := os.Chdir(root); err != nil {
		return fmt.Errorf("Error changing pwd to merged: %v", err)
	}
	pivotDir, err := ioutil.TempDir(root, ".pivot_root")
	if err != nil {
		return fmt.Errorf("can't create pivot_dir %v", err)
	}

	var mounted bool
	defer func() {
		if mounted {
			if errCleanup := syscall.Unmount(pivotDir, syscall.MNT_DETACH); errCleanup != nil {
				if err == nil {
					err = errCleanup
				}
				return
			}
		}

		errCleanup := os.Remove(pivotDir)
		if errCleanup != nil {
			errCleanup = fmt.Errorf("Error cleaning up after pivot_root: %v", errCleanup)
			if err == nil {
				err = errCleanup
			}
		}
	}()

	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		log.Infof("Try pivot %s failed, use chroot: %v", root, err)
		if err := os.Remove(pivotDir); err != nil {
			return fmt.Errorf("Error removing pivot_dir: %f", err)
		}
		return realChroot(root)
	}
	log.Infoln("pivot_root success!")
	mounted = true

	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("Error chdir to / after pivot: %v", err)
	}

	pivotDir = filepath.Join("/", filepath.Base(pivotDir))
	if err := syscall.Mount("", pivotDir, "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("Error making old root private after pivot: %v", err)
	}
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("Error while unmounting old root after pivot: %v", err)
	}
	mounted = false
	log.Infoln("Lazy umount old rootfs success!")

	return os.Remove(pivotDir)
}

func setupMount(volume string, rootURL string) error {
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Error getpwd: %v", err)
	}
	log.Infof("Now location is: %s", pwd)

	if volume != "" {
		volumeURLs := strings.Split(volume, ":")
		if len(volumeURLs) == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			if volumeURLs[0][0] != '/' {
				return fmt.Errorf("-v need absolute path")
			}
			if err := mountVolume(rootURL, volumeURLs); err != nil {
				return err
			}
			log.Infof("volume mounted success: %q", volumeURLs)
		} else {
			return fmt.Errorf("Invalid volume parameter input")
		}
	}
	pivotRoot(pwd)

	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	if err := syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), ""); err != nil {
		log.Errorf("Error mounting proc: %v", err)
	}
	// syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
	return nil
}

func mountVolume(rootURL string, volumeURLs []string) error {
	parentURL := volumeURLs[0]
	if exist, err := util.PathExist(parentURL); err != nil {
		return err
	} else if !exist {
		if err := os.Mkdir(parentURL, 0777); err != nil {
			return fmt.Errorf("Error: %v", err)
		}
	}
	containerURL := filepath.Join(rootURL, "merged", volumeURLs[1])
	if exist, err := util.PathExist(containerURL); err != nil {
		return err
	} else if !exist {
		if err := os.Mkdir(containerURL, 0777); err != nil {
			return fmt.Errorf("Error: %v", err)
		}
	}
	if err := syscall.Mount(parentURL, containerURL, "bind", syscall.MS_BIND, ""); err != nil {
		return fmt.Errorf("Error mounting volume: %v", err)
	}
	return nil
}
