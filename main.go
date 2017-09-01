package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"syscall"
)

const cgroupMemoryHierarchyMount = "/sys/fs/cgroup/memory"

func checkErr(err error) {
	if err != nil {
		log.Println(err.Error())
		log.Fatal(runtime.Caller(1))
	}
}

func main() {
	var err error

	// sig := make(chan os.Signal, 1)
	// signal.Notify(sig, syscall.SIGINT)
	// go func() {
	// }

	cmd := exec.Command("bash")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}

	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	err = cmd.Start()
	checkErr(err)
	limitPath := path.Join(cgroupMemoryHierarchyMount, "testmemorylimit")
	os.Mkdir(limitPath, 0755)
	ioutil.WriteFile(path.Join(limitPath, "memory.limit_in_bytes"), []byte("100m"), 0644)
	err = ioutil.WriteFile(path.Join(limitPath, "tasks"), []byte(strconv.Itoa(cmd.Process.Pid)), 0644)
	checkErr(err)
	cmd.Process.Wait()

	cmd = exec.Command("cgdelete", "memory:testmemorylimit")
	err = cmd.Run()
	checkErr(err)
}
