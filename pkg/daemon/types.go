package daemon

import (
	"sync"
	"time"
)

type Container struct {
	ID          string
	PID         int
	Image       string
	Command     []string
	NetworkInfo *NetworkInfo
	CreatedAt   time.Time
	Status      string
}

type NetworkInfo struct {
	IP            string
	Gateway       string
	Bridge        string
	HostVeth      string
	ContainerVeth string
}

type Daemon struct {
	containers map[string]*Container
	mu         sync.RWMutex
}
