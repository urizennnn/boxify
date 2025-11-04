package container

import (
	"log"
	"os"
	"syscall"
)

func CreateOverlayFS(containerID string) (error, string) {
	upperDir := "/var/lib/boxify/boxify-container/" + containerID + "/upper"
	workDir := "/var/lib/boxify/boxify-container/" + containerID + "/work"
	mergedDir := "/var/lib/boxify/boxify-container/" + containerID + "/merged"
	log.Printf("Creating overlay FS for container %s\n", containerID)

	err := os.MkdirAll(upperDir, 0o755)
	if err != nil {
		log.Printf("Error: error creating directory for upperDir %v\n", err)
		return err, ""
	}

	err = os.MkdirAll(workDir, 0o755)
	if err != nil {
		log.Printf("Error: error creating directory for workDir %v\n", err)
		return err, ""
	}

	err = os.MkdirAll(mergedDir, 0o755)
	if err != nil {
		log.Printf("Error: error creating directory for mergedDir %v\n", err)
		return err, ""
	}
	opts := "lowerdir=/var/lib/boxify/boxify-rootfs,upperdir=" + upperDir + ",workdir=" + workDir
	log.Printf("mounting %v\n", mergedDir)
	err = syscall.Mount("overlay", mergedDir, "overlay", 0, opts)
	if err != nil {
		log.Printf("Error: failed to mount %v\n", err)
		return err, ""
	}
	log.Printf("creating oldroot directory\n")
	oldRoot := mergedDir + "/.oldroot"
	err = os.MkdirAll(oldRoot, 0o700)
	if err != nil {
		log.Printf("Error: failed to create old root %v\n", err)
		return err, ""
	}

	log.Printf("done\n")

	return nil, mergedDir
}
