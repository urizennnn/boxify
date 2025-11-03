package network

import (
	"net"

	"github.com/vishvananda/netlink"
)

type IPManager struct {
	bridgeCIDR string
	gateway    net.IP
	nextIP     net.IP
	allocated  map[string]net.IP
}

type BridgeManager struct {
	defaultBridge  string
	bridgeIP       string
	bridgeInstance netlink.Link
	containerIps   map[string]string
}

type VethManager struct{
	veths map[string][2]string
}

type NetworkManager struct {
	bridgeManager *BridgeManager
	ipManager     *IPManager
	vethManager   *VethManager
}
