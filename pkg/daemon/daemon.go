package daemon

import (
	"fmt"
	"net/http"

	"github.com/urizennnn/boxify/pkg/daemon/handlers"
	"github.com/urizennnn/boxify/pkg/daemon/types"
)

func (m *Daemon) GetContainer(id string) (*types.Container, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	container, exists := m.containers[id]
	if !exists {
		return nil, fmt.Errorf("container not found")
	}
	return container, nil
}

func (d *Daemon) HandleCreateRequest(w http.ResponseWriter, r *http.Request) {
	handlers.HandleCreate(d, w, r)
}
