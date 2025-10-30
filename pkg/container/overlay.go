package container

import (
	"fmt"
	"os"
	"syscall"

	"github.com/google/uuid"
)

func CreateOverlayFS() (error,string) {
	containerID := uuid.New().String()
	upperDir := "/tmp/boxify-container/" + containerID + "/upper"
	workDir := "/tmp/boxify-container/" + containerID + "/work"
	mergedDir := "/tmp/boxify-container/" + containerID + "/merged"

	err := os.MkdirAll(upperDir, 0o755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: error creating directory for upperDir %v\n", err)
		return err,""
	}

	err = os.MkdirAll(workDir, 0o755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: error creating directory for workDir %v\n", err)
		return err,""
	}

	err = os.MkdirAll(mergedDir, 0o755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: error creating directory for mergedDir %v\n", err)
		return err,""
	}

	opts := fmt.Sprintf("lowerdir=/tmp/boxify-root,upperdir=%s,workdir=%s", upperDir, workDir)
	err = syscall.Mount("overlay", mergedDir, "overlay", 0, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: errour creating overlay FS %v\n", err)
		return err,""
	}

	syscall.PivotRoot(mergedDir, "/tmp/boxify-root/old")
	syscall.Chdir("/")
	defer syscall.Unmount(mergedDir, syscall.MNT_DETACH)
	defer os.RemoveAll("/tmp/boxify-container/" + containerID)

	return nil, containerID
}
