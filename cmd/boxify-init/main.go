package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/urizennnn/boxify/pkg/network"
)

func main() {
	if len(os.Args) < 7 {
		log.Fatalf("Usage: boxify-init <containerID> <memory> <cpu> <containerVeth> <gateway> <ipAddr>")
	}

	containerID := os.Args[1]
	containerVeth := os.Args[4]
	gateway := os.Args[5]
	ipAddr := os.Args[6]

	mergedDir := "/var/lib/boxify/boxify-container/" + containerID + "/merged"
	defer syscall.Unmount(mergedDir, syscall.MNT_DETACH)
	defer os.RemoveAll("/var/lib/boxify/boxify-container/" + containerID)
	defer os.RemoveAll("/sys/fs/cgroup/boxify/")

	setupMounts()

	if err := network.SetupContainerNetworkStandalone(containerID, containerVeth, gateway, ipAddr); err != nil {
		log.Fatalf("Error setting up network: %v\n", err)
	}

	cmd := exec.Command("/bin/sh")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Env = []string{"PATH=/bin:/usr/bin:/sbin:/usr/sbin"}

	if err := cmd.Run(); err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	log.Println("Container exiting...")
}

func setupMounts() {
	log.Printf("setting up proc mount\n")
	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		log.Fatalf("Error mounting proc: %v\n", err)
	}

	log.Printf("setting up sys mount\n")
	if err := syscall.Mount("sysfs", "/sys", "sysfs", 0, ""); err != nil {
		log.Fatalf("Error mounting sys: %v\n", err)
	}

	log.Printf("setting up dev mount\n")
	if err := syscall.Mount("tmpfs", "/dev", "tmpfs", 0, ""); err != nil {
		log.Fatalf("Error mounting dev: %v\n", err)
	}
}
