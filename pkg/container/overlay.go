package container

import (
	"fmt"
	"os"
	"syscall"

	"github.com/google/uuid"
)


func CreateOverlayFS() (error, string) {
	containerID := uuid.New().String()
	upperDir := "/tmp/boxify-container/" + containerID + "/upper"
	workDir := "/tmp/boxify-container/" + containerID + "/work"
	mergedDir := "/tmp/boxify-container/" + containerID + "/merged"
	fmt.Printf("Creating overlay FS for container %s\n", containerID)

	err := os.MkdirAll(upperDir, 0o755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: error creating directory for upperDir %v\n", err)
		return err, ""
	}

	err = os.MkdirAll(workDir, 0o755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: error creating directory for workDir %v\n", err)
		return err, ""
	}

	err = os.MkdirAll(mergedDir, 0o755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: error creating directory for mergedDir %v\n", err)
		return err, ""
	}
	opts := fmt.Sprintf("lowerdir=/tmp/boxify-rootfs,upperdir=%s,workdir=%s", upperDir, workDir)
	fmt.Printf("mounting %v\n", mergedDir)
	err = syscall.Mount("overlay", mergedDir, "overlay", 0, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to mount %v\n", err)
		return err, ""
	}
	fmt.Printf("creating oldroot directory\n")
	oldRoot := mergedDir + "/.oldroot"
	err = os.MkdirAll(oldRoot, 0o700)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create old root %v\n", err)
		return err, ""
	}

	fmt.Printf("pivoting\n")
	err = syscall.PivotRoot(mergedDir, oldRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: error in pivot function %v\n", err)
		return err, ""
	}
	fmt.Printf("changing dir\n")
	err = syscall.Chdir("/")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: error in pivot function %v\n", err)
		return err, ""
	}
	fmt.Printf("done\n")

	return nil, containerID
}
