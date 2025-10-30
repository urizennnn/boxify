package cgroup

import (
	"fmt"
	"os"
	"strconv"
)

func SetupCgroups(pid int) error{
	err := os.MkdirAll("/sys/fs/cgroup/boxify", 0o755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: error creating directory for cgroup %v\n", err)
		return
	}

	os.WriteFile("/sys/fs/cgroup/memory/boxify/memory.limit_in_bytes",
		[]byte("104857600"), 0o644)

	os.WriteFile("/sys/fs/cgroup/memory/boxify/tasks",
		[]byte(strconv.Itoa(pid)), 0o644)
}
