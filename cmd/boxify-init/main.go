package main

import (
	"log"
	"os"
	"syscall"
)

func main() {
	if len(os.Args) < 5 {
		log.Fatalf("Usage: boxify-init <containerID> <memory> <cpu> <mergedDir>")
	}

	containerID := os.Args[1]
	mergedDir := os.Args[4]
	err := syscall.Chroot(mergedDir)
	if err != nil {
		log.Printf("Error: error in pivot function %v\n", err)
		return
	}
	log.Printf("changing dir\n")
	err = syscall.Chdir("/")
	if err != nil {
		log.Printf("Error: error in pivot function %v\n", err)
		return
	}

	defer syscall.Unmount(mergedDir, syscall.MNT_DETACH)
	defer os.RemoveAll("/var/lib/boxify/boxify-container/" + containerID)
	defer os.RemoveAll("/sys/fs/cgroup/boxify/")

	setupMounts()

	log.Println("Container ready, waiting for attach...")
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
