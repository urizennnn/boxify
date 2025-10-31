package cgroup

import (
	"os"
	"strconv"
)

func SetupCgroupsV2(pid int) error {
	cgroupPath := "/sys/fs/cgroup/boxify"

	if err := os.MkdirAll(cgroupPath, 0o755); err != nil {
		return err
	}

	// Memory: 100MB
	os.WriteFile(cgroupPath+"/memory.max", []byte("104857600"), 0o644)

	// CPU: 50% (50000 out of 100000 microseconds)
	os.WriteFile(cgroupPath+"/cpu.max", []byte("50000 100000"), 0o644)

	// PIDs: max 100 processes
	os.WriteFile(cgroupPath+"/pids.max", []byte("100"), 0o644)

	os.WriteFile(cgroupPath+"/cgroup.procs",
		[]byte(strconv.Itoa(pid)), 0o644)

	return nil
}
