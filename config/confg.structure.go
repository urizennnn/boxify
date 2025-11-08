package config

import "github.com/urizennnn/boxify/pkg/daemon/types"

type ConfigStructure struct {
	ImageName string   `yaml:"image_name" json:"image_name"`
	Settings  Settings `yaml:"settings" json:"settings"`
}

type Settings struct {
	MemoryLimit string `yaml:"memory_limit" json:"memory_limit"`
	CpuLimit    string `yaml:"cpu_limit" json:"cpu_limit"`
}

type NetworkStorage struct {
	Id         string             `yaml:"id" json:"id"`
	Name       string             `yaml:"name" json:"name"`
	CreatedAt  string             `yaml:"created_at" json:"created_at"`
	Bridge     NetworkBridge      `yaml:"bridge" json:"bridge"`
	Ipam       NetworkIpam        `yaml:"ipam" json:"ipam"`
	Containers []*types.Container `yaml:"containers" json:"containers"`
}

type NetworkBridge struct {
	Name string `yaml:"name" json:"name"`
	Mtu  int    `yaml:"mtu" json:"mtu"`
}

type NetworkIpam struct {
	Subnet       string            `yaml:"subnet" json:"subnet"`
	Gateway      string            `yaml:"gateway" json:"gateway"`
	NextIP       string            `yaml:"next_ip" json:"next_ip"`
	AllocatedIPs map[string]string `yaml:"allocated_ips" json:"allocated_ips"`
}
