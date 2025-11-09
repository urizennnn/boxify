package container

import (
	"log"

	"github.com/urizennnn/boxify/pkg/daemon/types"
	"github.com/urizennnn/boxify/pkg/storage"
)

// ListAllContainers retrieves all containers from storage
func ListAllContainers() []*types.Container {
	repo := storage.NewNetworkRepository()
	containers, err := repo.ListAllContainers()
	if err != nil {
		log.Printf("Error listing containers: %v\n", err)
		return nil
	}
	return containers
}
