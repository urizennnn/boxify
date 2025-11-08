package network

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/urizennnn/boxify/config"
	"github.com/urizennnn/boxify/pkg/daemon/types"

	"gopkg.in/yaml.v3"
)

const NetworkStorageDir = "/var/lib/boxify/networks"

func WriteNetworkConfig(networkStorage *config.NetworkStorage) error {
	if err := os.MkdirAll(NetworkStorageDir, 0o755); err != nil {
		return fmt.Errorf("failed to create network storage directory: %w", err)
	}

	configPath := filepath.Join(NetworkStorageDir, "default.yaml")

	lock := NewFileLock(configPath)
	if err := lock.AcquireLock(); err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer lock.ReleaseLock()

	networkStorage.CreatedAt = time.Now().Format(time.RFC3339)

	data, err := yaml.Marshal(networkStorage)
	if err != nil {
		return fmt.Errorf("failed to marshal network config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write network config: %w", err)
	}

	return nil
}

func UpdateContainerInNetwork(networkId string, container *types.Container) error {
	configPath := filepath.Join(NetworkStorageDir, "default.yaml")

	lock := NewFileLock(configPath)
	if err := lock.AcquireLock(); err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer lock.ReleaseLock()

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read network config: %w", err)
	}

	var networkStorage config.NetworkStorage
	if err := yaml.Unmarshal(data, &networkStorage); err != nil {
		return fmt.Errorf("failed to unmarshal network config: %w", err)
	}

	container.CreatedAt = time.Now()
	networkStorage.Containers = append(networkStorage.Containers, container)

	updatedData, err := yaml.Marshal(&networkStorage)
	if err != nil {
		return fmt.Errorf("failed to marshal updated network config: %w", err)
	}

	if err := os.WriteFile(configPath, updatedData, 0o644); err != nil {
		return fmt.Errorf("failed to write updated network config: %w", err)
	}

	return nil
}

func ReadNetworkConfig(networkId string) (*config.NetworkStorage, error) {
	configPath := filepath.Join(NetworkStorageDir, "default.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read network config: %w", err)
	}

	var networkStorage config.NetworkStorage
	if err := yaml.Unmarshal(data, &networkStorage); err != nil {
		return nil, fmt.Errorf("failed to unmarshal network config: %w", err)
	}

	return &networkStorage, nil
}

func CheckNetworkConfigExists() bool {
	configPath := filepath.Join(NetworkStorageDir, "default.yaml")
	info, err := os.Stat(configPath)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
