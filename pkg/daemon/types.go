package daemon

import (
	"fmt"
	"sync"

	"github.com/urizennnn/boxify/pkg/daemon/types"
	"github.com/urizennnn/boxify/pkg/network"
)

type Daemon struct {
	containers map[string]*types.Container
	mu         sync.RWMutex
	networkMgr *network.NetworkManager
}

func New() *Daemon {
	networkMgr, err := network.NewNetworkManager()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize network manager: %v", err))
	}

	return &Daemon{
		containers: make(map[string]*types.Container),
		networkMgr: networkMgr,
	}
}

func (d *Daemon) AddContainer(container *types.Container) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.containers[container.ID] = container
}

func (d *Daemon) NetworkManager() *network.NetworkManager {
	return d.networkMgr
}
