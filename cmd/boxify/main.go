package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/urizennnn/boxify/pkg/container"

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
	err, _ := container.InitContainer()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: errour creating overlay FS %v\n", err)
		os.Exit(1)
	}
	cmd := exec.Command("/bin/sh")
	setupMounts()
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Env = []string{"PATH=/bin:/usr/bin:/sbin:/usr/sbin"}
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Container exiting...")
}

func setupMounts() {
	err := syscall.Unmount("/proc", syscall.MNT_DETACH)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	err = syscall.Mount("proc", "/proc", "proc", 0, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	err = syscall.Mount("sysfs", "/sys", "sysfs", 0, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	err = syscall.Mount("tmpfs", "/dev", "tmpfs", 0, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
