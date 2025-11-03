package daemon

import "fmt"

func (m *Daemon) GetContainer(id string) (*Container, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	container, exists := m.containers[id]
	if !exists {
		return nil, fmt.Errorf("container not found")
	}
	return container, nil
}
