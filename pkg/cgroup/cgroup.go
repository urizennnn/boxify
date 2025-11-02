package cgroup

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func SetupCgroupsV2(pid int, mem, cpu string) error {
	cgroupPath := "/sys/fs/cgroup/boxify"

	if err := os.MkdirAll(cgroupPath, 0o755); err != nil {
		return err
	}

	calculatedMem, err := parseMemory(strings.ToLower(mem))
	if err != nil {
		return err
	}
	err = os.WriteFile(cgroupPath+"/memory.max", fmt.Appendf(nil, "%d", calculatedMem), 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: error setting memory limit %v\n", err)
		return err
	}

	calculatedCpu, err := strconv.Atoi(cpu)
	if err != nil {
		return err
	}
	calculatedCpu *= 1000
	err = os.WriteFile(cgroupPath+"/cpu.max", fmt.Appendf(nil, "%d %d", calculatedCpu, 100000), 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: error setting CPU limit %v\n", err)
		return err
	}

	// PIDs: max 100 processes
	err = os.WriteFile(cgroupPath+"/pids.max", []byte("100"), 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: error setting PIDs limit %v\n", err)
		return err
	}

	err = os.WriteFile(cgroupPath+"/cgroup.procs",
		[]byte(strconv.Itoa(pid)), 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: error adding pid to cgroup %v\n", err)
		return err
	}

	return nil
}
func parseMemory(input string) (int64, error) {
    value, _ := strconv.ParseInt(input[:len(input)-1], 10, 64)
    
    switch input[len(input)-1] {
    case 'k', 'K':
        return value * 1024, nil
    case 'm', 'M':
        return value * 1024 * 1024, nil
    case 'g', 'G':
        return value * 1024 * 1024 * 1024, nil
    }
    
    return 0, fmt.Errorf("invalid format")
}

