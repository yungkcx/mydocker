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
)

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
	// Mount an overlayfs for the container.
	if err := syscall.Mount("overlay", filepath.Join(root, "merged"), "overlay", 0, "upperdir=./upper,lowerdir=./root,workdir=./work"); err != nil {
		return fmt.Errorf("Error creating overlayfs: %v", err)
	}
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

	return os.Remove(pivotDir)
}

func setupMount() error {
	if err := os.Chdir("./overlay/container"); err != nil {
		return fmt.Errorf("Error chdir to container: %v", err)
	}
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Error getpwd: %v", err)
	}
	log.Infof("Now location is: %s", pwd)

	pivotRoot(pwd)

	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	if err := syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), ""); err != nil {
		log.Errorf("Error mounting proc: %v", err)
	}
	// syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
	return nil
}

// RunContainerInitProcess is a function.
func RunContainerInitProcess() error {
	args := readUserCommand()
	if err := setupMount(); err != nil {
		return err
	}
	path, err := exec.LookPath(args[0])
	if err != nil {
		return fmt.Errorf("Error lookpath %s: %v", args[0], err)
	}
	if err := syscall.Exec(path, args, nil); err != nil {
		return fmt.Errorf("Error exec container command: %v", err)
	}
	return nil
}
