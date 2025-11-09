package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/urizennnn/boxify/pkg/daemon/types"
	"github.com/urizennnn/boxify/pkg/network"

	"gopkg.in/yaml.v3"
)

const NetworkStorageDir = "/var/lib/boxify/networks"

// NetworkRepository handles persistence operations for network data
type NetworkRepository struct {
	storageDir string
}

// NewNetworkRepository creates a new network repository
func NewNetworkRepository() *NetworkRepository {
	return &NetworkRepository{
		storageDir: NetworkStorageDir,
	}
}

// WriteNetworkConfig persists network storage to disk
func (r *NetworkRepository) WriteNetworkConfig(networkStorage *network.NetworkStorage) error {
	if err := os.MkdirAll(r.storageDir, 0o755); err != nil {
		return fmt.Errorf("failed to create network storage directory: %w", err)
	}

	configPath := filepath.Join(r.storageDir, "default.yaml")

	lock := network.NewFileLock(configPath)
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

// UpdateContainerInNetwork adds or updates a container in the network storage
func (r *NetworkRepository) UpdateContainerInNetwork(networkId string, container *types.Container) error {
	configPath := filepath.Join(r.storageDir, "default.yaml")

	lock := network.NewFileLock(configPath)
	if err := lock.AcquireLock(); err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer lock.ReleaseLock()

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read network config: %w", err)
	}

	var networkStorage network.NetworkStorage
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

// ReadNetworkConfig reads network storage from disk
func (r *NetworkRepository) ReadNetworkConfig(networkId string) (*network.NetworkStorage, error) {
	configPath := filepath.Join(r.storageDir, "default.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read network config: %w", err)
	}

	var networkStorage network.NetworkStorage
	if err := yaml.Unmarshal(data, &networkStorage); err != nil {
		return nil, fmt.Errorf("failed to unmarshal network config: %w", err)
	}

	return &networkStorage, nil
}

// CheckNetworkConfigExists checks if network config file exists
func (r *NetworkRepository) CheckNetworkConfigExists() bool {
	configPath := filepath.Join(r.storageDir, "default.yaml")
	info, err := os.Stat(configPath)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// ListAllContainers retrieves all containers from network storage
func (r *NetworkRepository) ListAllContainers() ([]*types.Container, error) {
	configPath := filepath.Join(r.storageDir, "default.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read network config: %w", err)
	}

	var networkStorage network.NetworkStorage
	if err := yaml.Unmarshal(data, &networkStorage); err != nil {
		return nil, fmt.Errorf("failed to unmarshal network config: %w", err)
	}

	return networkStorage.Containers, nil
}
