package network

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
)

func (m *BridgeManager) CreateBridgeWithIp(ip *IPManager) error {
	la := netlink.NewLinkAttrs()
	la.Name = "boxify-bridge0"
	m.defaultBridge = la.Name

	existingLink, err := LinkExists(la.Name)
	if err == nil {
		bridge, ok := existingLink.(*netlink.Bridge)
		if ok {
			m.bridgeInstance = bridge
		} else {
			boxifyBridge := &netlink.Bridge{LinkAttrs: la}
			m.bridgeInstance = boxifyBridge
		}
	} else {
		boxifyBridge := &netlink.Bridge{LinkAttrs: la}
		m.bridgeInstance = boxifyBridge
		err := netlink.LinkAdd(boxifyBridge)
		if err != nil {
			fmt.Printf("could not add %s: %v\n", la.Name, err)
			return err
		}
	}

	err = netlink.LinkSetUp(m.bridgeInstance)
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

	err = netlink.AddrAdd(m.bridgeInstance, addr)
	if err != nil {
		fmt.Printf("could not add ip addr to bridge, %v", err)
		return err
	}
	ip.nextIP = ip.IncrementIp(string(ip.nextIP))
	ip.allocated[la.Name] = net.IP(nonConflictingAddr)

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
