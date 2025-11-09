package network

import (
	"github.com/urizennnn/boxify/pkg/daemon/types"
	"github.com/urizennnn/boxify/pkg/storage"
)

const NetworkStorageDir = "/var/lib/boxify/networks"

var defaultRepo = storage.NewNetworkRepository()

// WriteNetworkConfig persists network storage to disk
// Deprecated: Use storage.NetworkRepository directly
func WriteNetworkConfig(networkStorage *NetworkStorage) error {
	return defaultRepo.WriteNetworkConfig(networkStorage)
}

// UpdateContainerInNetwork adds or updates a container in the network storage
// Deprecated: Use storage.NetworkRepository directly
func UpdateContainerInNetwork(networkId string, container *types.Container) error {
	return defaultRepo.UpdateContainerInNetwork(networkId, container)
}

// ReadNetworkConfig reads network storage from disk
// Deprecated: Use storage.NetworkRepository directly
func ReadNetworkConfig(networkId string) (*NetworkStorage, error) {
	return defaultRepo.ReadNetworkConfig(networkId)
}

// CheckNetworkConfigExists checks if network config file exists
// Deprecated: Use storage.NetworkRepository directly
func CheckNetworkConfigExists() bool {
	return defaultRepo.CheckNetworkConfigExists()
}
