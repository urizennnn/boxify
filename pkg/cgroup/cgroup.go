package cgroup

import (
	"log"
	"os"
	"strconv"
	"strings"
)

func SetupCgroupsV2(pid int, mem, cpu string) error {
	log.Printf("Setting up cgroups v2 with memory: %s, cpu: %s for pid: %d\n", mem, cpu, pid)
	cgroupPath := "/sys/fs/cgroup/boxify"

	if err := os.MkdirAll(cgroupPath, 0o755); err != nil {
		return err
	}

	calculatedMem, err := parseMemory(strings.ToLower(mem))
	if err != nil {
		return err
	}
	err = os.WriteFile(cgroupPath+"/memory.max", []byte(strconv.FormatInt(calculatedMem, 10)), 0o644)
	if err != nil {
		log.Printf("Error: error setting memory limit %v\n", err)
		return err
	}

	calculatedCpu, err := strconv.Atoi(cpu)
	if err != nil {
		return err
	}
	calculatedCpu *= 1000
	err = os.WriteFile(cgroupPath+"/cpu.max", []byte(strconv.Itoa(calculatedCpu)+" 100000"), 0o644)
	if err != nil {
		log.Printf("Error: error setting CPU limit %v\n", err)
		return err
	}

	err = os.WriteFile(cgroupPath+"/pids.max", []byte("100"), 0o644)
	if err != nil {
		log.Printf("Error: error setting PIDs limit %v\n", err)
		return err
	}

	err = os.WriteFile(cgroupPath+"/cgroup.procs",
		[]byte(strconv.Itoa(pid)), 0o644)
	if err != nil {
		log.Printf("Error: error adding pid to cgroup %v\n", err)
		return err
	}

	return nil
}
func parseMemory(input string) (int64, error) {
	if input == "" {
		return 0, nil
	}

	if len(input) < 2 {
		value, err := strconv.ParseInt(input, 10, 64)
		if err != nil {
			return 0, err
		}
		return value, nil
	}

	lastChar := input[len(input)-1]
	value, err := strconv.ParseInt(input[:len(input)-1], 10, 64)
	if err != nil {
		return 0, err
	}

	switch lastChar {
	case 'k', 'K':
		return value * 1024, nil
	case 'm', 'M':
		return value * 1024 * 1024, nil
	case 'g', 'G':
		return value * 1024 * 1024 * 1024, nil
	default:
		valueWithoutSuffix, err := strconv.ParseInt(input, 10, 64)
		if err != nil {
			return 0, err
		}
		return valueWithoutSuffix, nil
	}
}

