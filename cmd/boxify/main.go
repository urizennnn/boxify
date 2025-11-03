package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/urizennnn/boxify/pkg/cgroup"
	"github.com/urizennnn/boxify/pkg/container"
	syscall "golang.org/x/sys/unix"
)

func main() {
	memory := "100m"
	cpu := "50"

	for i, arg := range os.Args {
		if arg == "--memory" && i+1 < len(os.Args) {
			memory = os.Args[i+1]
		}
		if arg == "--cpu" && i+1 < len(os.Args) {
			cpu = os.Args[i+1]
		}
	}
	switch os.Args[1] {
	case "run":
		parent(memory, cpu)
	case "child":
		child()
	}
}

func parent(memory, cpu string) {
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

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	pid := cmd.Process.Pid
	cgroup.SetupCgroupsV2(pid, memory, cpu)
	cmd.Wait()
}

func child() {
	err, containerID := container.InitContainer()

	mergedDir := "/tmp/boxify-container/" + containerID + "/merged"
	defer syscall.Unmount(mergedDir, syscall.MNT_DETACH)
	defer os.RemoveAll("/tmp/boxify-container/" + containerID)
	defer os.RemoveAll("/sys/fs/cgroup/boxify/")

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed in creating overlay FS %v\n", err)
		os.Exit(1)
	}
	setupMounts()

	cmd := exec.Command("/bin/sh")
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
	fmt.Printf("setting up proc mount\n")
	err := syscall.Mount("proc", "/proc", "proc", 0, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("setting up sys mount\n")
	err = syscall.Mount("sysfs", "/sys", "sysfs", 0, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("setting up dev mount\n")
	err = syscall.Mount("tmpfs", "/dev", "tmpfs", 0, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
