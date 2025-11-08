package main

import (
	"log"
	"os"
	"path/filepath"
	"syscall"
)

func main() {
	if len(os.Args) < 5 {
		log.Fatalf("Usage: boxify-init <containerID> <memory> <cpu> <mergedDir>")
	}

	containerID := os.Args[1]
	mergedDir := os.Args[4]

	if err := pivotRoot(mergedDir); err != nil {
		log.Fatalf("Error: failed to pivot root: %v\n", err)
	}

	defer os.RemoveAll("/var/lib/boxify/boxify-container/" + containerID)
	defer os.RemoveAll("/sys/fs/cgroup/boxify/")

	setupMounts()

	log.Println("Container ready, waiting for attach...")

	for {
		syscall.Pause()
	}
}

func pivotRoot(newRoot string) error {
	log.Printf("pivoting root to %s\n", newRoot)

	putOld := filepath.Join(newRoot, ".pivot_root")
	if err := os.MkdirAll(putOld, 0700); err != nil {
		return err
	}

	if err := syscall.Mount(newRoot, newRoot, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return err
	}

	if err := syscall.PivotRoot(newRoot, putOld); err != nil {
		return err
	}

	if err := os.Chdir("/"); err != nil {
		return err
	}

	putOld = "/.pivot_root"
	if err := syscall.Unmount(putOld, syscall.MNT_DETACH); err != nil {
		return err
	}

	if err := os.RemoveAll(putOld); err != nil {
		return err
	}

	log.Printf("successfully pivoted root\n")
	return nil
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
