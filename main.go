package main

import (
	"fmt"
	"os"
	"os/exec"

	syscall "golang.org/x/sys/unix"
)

func main() {
	switch os.Args[1] {
	case "run":
		parent()
	case "child":
		child()
	}
}

func parent() {
	cmd := exec.Command("/proc/self/exe", "child")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWNS,

		Unshareflags: syscall.CLONE_NEWNS,
	}
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func child() {
	syscall.Chroot("/tmp/boxify-root")
	syscall.Chdir("/")

	cmd := exec.Command("/bin/sh")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
cmd.Env = []string{"PATH=/bin:/usr/bin:/sbin:/usr/sbin"}
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
