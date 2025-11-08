package types

import (
	"os/exec"
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
	Cmd         *exec.Cmd `yaml:"-"`
}

type NetworkInfo struct {
	IP            string
	Gateway       string
	Bridge        string
	HostVeth      string
	ContainerVeth string
}
