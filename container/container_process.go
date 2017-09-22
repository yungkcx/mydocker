package container

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	log "github.com/Sirupsen/logrus"
)

// NewPipe returns a pipe.
func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, nil
}

// NewParentProcess use init command to init container.
func NewParentProcess(tty bool, volume string, rootURL string) (*exec.Cmd, *os.File) {
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		log.Errorf("New pipe error %v", err)
		return nil, nil
	}

	// Because of syscall.Unshare() can't work, I use this.
	cmd := exec.Command("unshare", "-m", "/home/yungkc/Golang/bin/mydocker", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	cmd.ExtraFiles = []*os.File{readPipe}
	if err := NewWorkSpace(rootURL); err != nil {
		log.Errorf(err.Error())
		return nil, nil
	}
	cmd.Dir = rootURL

	return cmd, writePipe
}

// NewWorkSpace create a new volume mount point for the container.
func NewWorkSpace(rootURL string) error {
	if err := createReadonlyLayer(rootURL); err != nil {
		return err
	}
	if err := createWriteLayer(rootURL); err != nil {
		return err
	}
	if err := createMountPoint(rootURL); err != nil {
		return err
	}
	return nil
}

func createMountPoint(rootURL string) error {
	// Mount an overlayfs for the container.
	mergedDir := filepath.Join(rootURL, "merged")
	if err := os.Mkdir(mergedDir, 0777); err != nil && os.IsNotExist(err) {
		return fmt.Errorf("Error: %v", err)
	}
	upperdir := filepath.Join(rootURL, "upper")
	lowerdir := filepath.Join(rootURL, "busybox")
	workdir := filepath.Join(rootURL, "work")
	dirs := "upperdir=" + upperdir + ",lowerdir=" + lowerdir + ",workdir=" + workdir
	if err := syscall.Mount("overlay", mergedDir, "overlay", 0, dirs); err != nil {
		return fmt.Errorf("Error mounting overlayfs: %v", err)
	}
	return nil
}

func createWriteLayer(rootURL string) error {
	writeURL := filepath.Join(rootURL, "upper")
	if err := os.Mkdir(writeURL, 0777); err != nil && os.IsNotExist(err) {
		return fmt.Errorf("Error: %v", err)
	}
	workdir := filepath.Join(rootURL, "work")
	if err := os.Mkdir(workdir, 0777); err != nil && os.IsNotExist(err) {
		return fmt.Errorf("Error: %v", err)
	}
	return nil
}

func createReadonlyLayer(rootURL string) error {
	busyboxURL := filepath.Join(rootURL, "busybox")
	if _, err := pathExist(busyboxURL); err != nil {
		return err
	}
	return nil
}

func pathExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("Error judging whether %s exists: %v", path, err)
}

// DeleteWorkSpace delete the overlayfs workspace.
func DeleteWorkSpace(volume string, rootURL string) error {
	if err := deleteMountPoint(rootURL); err != nil {
		return err
	}
	if err := deleteWriteLayer(rootURL); err != nil {
		return err
	}
	return nil
}

func deleteMountPoint(rootURL string) error {
	mergedDir := filepath.Join(rootURL, "merged")
	if err := syscall.Unmount(mergedDir, syscall.MNT_FORCE); err != nil {
		return fmt.Errorf("Error umounting overlayfs: %v", err)
	}
	if err := os.RemoveAll(mergedDir); err != nil {
		return fmt.Errorf("Error removing %s: %v", mergedDir, err)
	}
	return nil
}

func deleteWriteLayer(rootURL string) error {
	if err := os.RemoveAll(filepath.Join(rootURL, "upper")); err != nil {
		return fmt.Errorf("Error removing upperdir: %v", err)
	}
	if err := os.RemoveAll(filepath.Join(rootURL, "work")); err != nil {
		return fmt.Errorf("Error removing workdir: %v", err)
	}
	return nil
}
