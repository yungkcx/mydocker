package main

import (
	"log"
	"os"
	"os/exec"
	"runtime"
	"syscall"
)

func checkErr(err error) {
	if err != nil {
		log.Println(err.Error())
		log.Fatal(runtime.Caller(1))
	}
}

func main() {
	cmd := exec.Command("zsh")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWUSER | syscall.CLONE_NEWNET,
	}

	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	err := cmd.Run()
	checkErr(err)
}
