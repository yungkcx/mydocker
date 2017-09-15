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
	// log.Infoln(os.Getpid())
	// if err := syscall.Mount(root, root, "", syscall.MS_BIND, ""); err != nil {
	// 	return fmt.Errorf("Error mounting pivot dir before pivot: %v", err)
	// }
	// if err := syscall.Mount("overlay", "./merged", "overlay", 0, "upperdir=./upper,lowerdir=./root,workdir=./work"); err != nil {
	// 	return fmt.Errorf("Error creating overlay: %v", err)
	// }
	// if err := os.Chdir("./merged"); err != nil {
	// 	return fmt.Errorf("Error changing pwd to merged: %v", err)
	// }

	if err := syscall.Unshare(syscall.CLONE_NEWNS); err != nil {
		return fmt.Errorf("Error creating mount namespace before pivot: %v", err)
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
	mounted = true

	pivotDir = filepath.Join("/", filepath.Base(pivotDir))
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("Error changing to new root: %v", err)
	}
	if err := syscall.Mount("", pivotDir, "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("Error making old root private after pivot: %v", err)
	}
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("Error while unmounting old root after pivot: %v", err)
	}
	mounted = false

	return os.Remove(pivotDir)
}

func setupMount() {
	pwd, err := os.Getwd()
	if err != nil {
		log.Errorf("Get current location error %v", err)
		return
	}
	log.Infof("Current location is %s", pwd)

	pivotRoot(pwd)

	// defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	// syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	// syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
}

// RunContainerInitProcess is a function.
func RunContainerInitProcess() error {
	cmdArray := readUserCommand()
	if len(cmdArray) == 0 {
		return fmt.Errorf("Run container get user command error, cmdArray is nil")
	}

	setupMount()

	path, err := exec.LookPath(cmdArray[0])
	// path := filepath.Join("/bin", cmdArray[0])
	if err != nil {
		log.Errorf("Exec look path error %v", err)
		return err
	}
	log.Infof("Found path %s", path)
	if err := syscall.Exec(path, cmdArray, nil /*os.Environ()*/); err != nil {
		log.Errorf(err.Error())
	}
	return nil
}
