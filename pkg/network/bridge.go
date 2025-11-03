package network

import (
	"fmt"

	"github.com/vishvananda/netlink"
)

func (m *BridgeManager) CreateBridgeWithIp(ip *IPManager) error {
	la := netlink.NewLinkAttrs()
	la.Name = "boxify-bridge"
	m.defaultBridge = la.Name
	boxifyBridge := &netlink.Bridge{LinkAttrs: la}
	m.bridgeInstance = boxifyBridge
	err := netlink.LinkAdd(boxifyBridge)
	if err != nil {
		fmt.Printf("could not add %s: %v\n", la.Name, err)
		return err
	}

	err = netlink.LinkSetUp(boxifyBridge)
	if err != nil {
		fmt.Printf("could not set up %s: %v\n", la.Name, err)
		return err
	}
	nonConflictingAddr := ip.GetNextIP()
	addr, err := netlink.ParseAddr(nonConflictingAddr)
	if err != nil {
		fmt.Printf("failed to parse addr, %v\n", err)
		return err
	}

	err = netlink.AddrAdd(boxifyBridge, addr)
	if err != nil {
		fmt.Printf("could not add ip addr to bridge, %v", err)
		return err
	}
	ip.nextIP = ip.IncrementIp(ip.GetGateway())
	ip.allocated[la.Name] = ip.gateway

	return nil
}

func (m *BridgeManager) AttachIpToBridge(ipAddr string) error {
	bridgeLink, err := netlink.LinkByName(m.defaultBridge)
	if err != nil {
		fmt.Printf("could not find bridge %s: %v\n", m.defaultBridge, err)
		return err
	}
	addr, err := netlink.ParseAddr(ipAddr)
	if err != nil {
		fmt.Printf("failed to parse addr, %v\n", err)
		return err
	}
	err = netlink.AddrAdd(bridgeLink, addr)
	if err != nil {
		fmt.Printf("could not add ip addr to bridge, %v", err)
		return err
	}
	return nil
}
func (m *BridgeManager) ReturnBridgeDetails() *BridgeManager {
	return m
}

func (m *BridgeManager) BringDownBridge() error {
	bridgeLink, err := netlink.LinkByName(m.defaultBridge)
	if err != nil {
		fmt.Printf("could not find bridge %s: %v\n", m.defaultBridge, err)
		return err
	}
	err = netlink.LinkSetDown(bridgeLink)
	if err != nil {
		fmt.Printf("could not bring bridge down %s: %v\n", m.defaultBridge, err)
		return err
	}
	return nil
}


func (m *BridgeManager) BringUpBridge() error {
	bridgeLink, err := netlink.LinkByName(m.defaultBridge)
	if err != nil {
		fmt.Printf("could not find bridge %s: %v\n", m.defaultBridge, err)
		return err
	}
	err = netlink.LinkSetUp(bridgeLink)
	if err != nil {
		fmt.Printf("could not bring bridge up %s: %v\n", m.defaultBridge, err)
		return err
	}
	return nil
}



func (m *BridgeManager) DeleteBridge() error {
	bridgeLink, err := netlink.LinkByName(m.defaultBridge)
	if err != nil {
		fmt.Printf("could not find bridge %s: %v\n", m.defaultBridge, err)
		return err
	}
	err = netlink.LinkDel(bridgeLink)
	if err != nil {
		fmt.Printf("could not delete bridge %s: %v\n", m.defaultBridge, err)
		return err
	}
	return nil
}

