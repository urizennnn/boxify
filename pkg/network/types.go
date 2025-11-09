package network

import (
	"net"

	"github.com/urizennnn/boxify/pkg/daemon/types"
	"github.com/vishvananda/netlink"
)

type IPManager struct {
	BridgeCIDR string
	Gateway    net.IP
	NextIP     net.IP
	Allocated  map[string]net.IP
}

type BridgeManager struct {
	DefaultBridge  string
	BridgeIP       string
	BridgeInstance netlink.Link
	ContainerIps   map[string]string
	Status         string
}

type VethManager struct {
	veths map[string][2]string
}

type NatManager struct {
	BridgeManager *BridgeManager
	IpManager     *IPManager
}

type NetworkManager struct {
	BridgeManager *BridgeManager
	IpManager     *IPManager
	VethManager   *VethManager
	NatManager    *NatManager
}

// NetworkStorage represents the persisted state of a network and its containers
type NetworkStorage struct {
	Id         string             `yaml:"id" json:"id"`
	Name       string             `yaml:"name" json:"name"`
	CreatedAt  string             `yaml:"created_at" json:"created_at"`
	Bridge     NetworkBridge      `yaml:"bridge" json:"bridge"`
	Ipam       NetworkIpam        `yaml:"ipam" json:"ipam"`
	Containers []*types.Container `yaml:"containers" json:"containers"`
}

// NetworkBridge represents the bridge configuration for a network
type NetworkBridge struct {
	Name string `yaml:"name" json:"name"`
	Mtu  int    `yaml:"mtu" json:"mtu"`
}

// NetworkIpam represents IP address management configuration for a network
type NetworkIpam struct {
	Subnet       string            `yaml:"subnet" json:"subnet"`
	Gateway      string            `yaml:"gateway" json:"gateway"`
	NextIP       string            `yaml:"next_ip" json:"next_ip"`
	AllocatedIPs map[string]string `yaml:"allocated_ips" json:"allocated_ips"`
}
