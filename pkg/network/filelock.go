package network

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type FileLock struct {
	lockFilePath string
}

func NewFileLock(filePath string) *FileLock {
	return &FileLock{
		lockFilePath: filePath + ".lock",
	}
}

func (fl *FileLock) AcquireLock() error {
	if _, err := os.Stat(fl.lockFilePath); err == nil {
		pidBytes, readErr := os.ReadFile(fl.lockFilePath)
		if readErr != nil {
			return fmt.Errorf("lock file exists but cannot read PID: %w", readErr)
		}
		pidStr := strings.TrimSpace(string(pidBytes))
		return fmt.Errorf("file is locked by process %s", pidStr)
	}

	pid := os.Getpid()
	err := os.WriteFile(fl.lockFilePath, []byte(strconv.Itoa(pid)), 0644)
	if err != nil {
		return fmt.Errorf("failed to create lock file: %w", err)
	}

	return nil
}

func (fl *FileLock) ReleaseLock() error {
	err := os.Remove(fl.lockFilePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to release lock: %w", err)
	}
	return nil
}

func (fl *FileLock) IsLocked() bool {
	_, err := os.Stat(fl.lockFilePath)
	return err == nil
}
