package network

import (
	"net"

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
