package network

import (
	"log"
	"net"

	"github.com/google/uuid"
	"github.com/urizennnn/boxify/pkg/daemon/types"
	"github.com/vishvananda/netlink"
)

func (m *BridgeManager) CreateBridgeWithIp(ip *IPManager) error {
	la := netlink.NewLinkAttrs()
	la.Name = "boxify-bridge0"
	m.DefaultBridge = la.Name
	log.Printf("[1/8] Starting bridge creation for %s", la.Name)

	existingLink, err := LinkExists(la.Name)
	if err == nil {
		log.Printf("[2/8] Existing link found: %s", la.Name)
		bridge, ok := existingLink.(*netlink.Bridge)
		if ok {
			log.Printf("[3/8] Link is already a bridge instance, reusing it")
			m.BridgeInstance = bridge
		} else {
			log.Printf("[3/8] Link exists but is not a bridge, creating new bridge instance")
			boxifyBridge := &netlink.Bridge{LinkAttrs: la}
			m.BridgeInstance = boxifyBridge
		}
	} else {
		log.Printf("[2/8] No existing link found, creating new bridge %s", la.Name)
		boxifyBridge := &netlink.Bridge{LinkAttrs: la}
		m.BridgeInstance = boxifyBridge
		err := netlink.LinkAdd(boxifyBridge)
		if err != nil {
			log.Printf("[ERROR] could not add bridge %s: %v", la.Name, err)
			return err
		}
		log.Printf("[3/8] Bridge %s successfully created", la.Name)
	}

	log.Printf("[4/8] Bringing up bridge %s", la.Name)
	err = netlink.LinkSetUp(m.BridgeInstance)
	if err != nil {
		log.Printf("[ERROR] could not set up %s: %v", la.Name, err)
		return err
	}
	log.Printf("[5/8] Bridge %s is up", la.Name)

	var bridgeIP string
	existingAddrs, err := netlink.AddrList(m.BridgeInstance, netlink.FAMILY_V4)
	if err == nil && len(existingAddrs) > 0 {
		bridgeIP = existingAddrs[0].IP.String()
		log.Printf("[6/8] Bridge already has IP: %s, reusing it", bridgeIP)
		ip.Gateway = net.ParseIP(bridgeIP)
		ip.Allocated[la.Name] = net.ParseIP(bridgeIP)
		log.Printf("[8/8] Bridge %s setup complete with existing IP %s (Gateway set to %s)", la.Name, bridgeIP, ip.Gateway)
	} else {
		bridgeIP = ip.GetNextIP()
		log.Printf("[6/8] Got next available IP: %s", bridgeIP)
		addr, err := netlink.ParseAddr(bridgeIP + ip.BridgeCIDR)
		if err != nil {
			log.Printf("[ERROR] failed to parse IP addr %s: %v", bridgeIP, err)
			return err
		}

		log.Printf("[7/8] Assigning IP %s to bridge %s", addr, la.Name)
		err = netlink.AddrAdd(m.BridgeInstance, addr)
		if err != nil {
			log.Printf("[ERROR] could not add IP addr to bridge %s: %v", la.Name, err)
			return err
		}

		ip.Gateway = net.ParseIP(bridgeIP)
		incrementedIP := ip.IncrementIp(ip.NextIP.String())
		ip.NextIP = incrementedIP
		ip.Allocated[la.Name] = net.ParseIP(bridgeIP)
		log.Printf("[8/8] Bridge %s setup complete with IP %s (Gateway set to %s)", la.Name, bridgeIP, ip.Gateway)
	}

	if CheckNetworkConfigExists() {
		log.Printf("[INFO] Network config already exists, updating allocated IPs")
		configPath := NetworkStorageDir + "/default.yaml"
		lock := NewFileLock(configPath)
		if err := lock.AcquireLock(); err != nil {
			log.Printf("[WARNING] Failed to acquire lock: %v", err)
			return err
		}
		defer lock.ReleaseLock()

		networkStorage, err := ReadNetworkConfig("default")
		if err != nil {
			log.Printf("[WARNING] Failed to read network config: %v", err)
			return err
		}

		networkStorage.Ipam.AllocatedIPs[la.Name] = bridgeIP
		networkStorage.Ipam.NextIP = ip.NextIP.String()

		if err := WriteNetworkConfigWithoutLock(networkStorage); err != nil {
			log.Printf("[WARNING] Failed to update network config: %v", err)
			return err
		}
		log.Printf("[SUCCESS] Network config updated")
	} else {
		log.Printf("[INFO] Creating new network config")
		networkStorage := &NetworkStorage{
			Id:   uuid.New().String(),
			Name: la.Name,
			Bridge: NetworkBridge{
				Name: la.Name,
				Mtu:  m.BridgeInstance.Attrs().MTU,
			},
			Ipam: NetworkIpam{
				Subnet:       ip.BridgeCIDR,
				Gateway:      ip.Gateway.String(),
				NextIP:       ip.NextIP.String(),
				AllocatedIPs: make(map[string]string),
			},
			Containers: []*types.Container{},
		}

		for name, ipAddr := range ip.Allocated {
			networkStorage.Ipam.AllocatedIPs[name] = ipAddr.String()
		}

		if err := WriteNetworkConfig(networkStorage); err != nil {
			log.Printf("[WARNING] Failed to persist network config: %v", err)
		} else {
			log.Printf("[SUCCESS] Network config persisted to %s/default.yaml", NetworkStorageDir)
		}
	}

	return nil
}

func (m *BridgeManager) AttachIpToBridge(ipAddr string) error {
	bridgeLink, err := netlink.LinkByName(m.DefaultBridge)
	if err != nil {
		log.Printf("could not find bridge %s: %v\n", m.DefaultBridge, err)
		return err
	}
	addr, err := netlink.ParseAddr(ipAddr)
	if err != nil {
		log.Printf("failed to parse addr, %v\n", err)
		return err
	}
	err = netlink.AddrAdd(bridgeLink, addr)
	if err != nil {
		log.Printf("could not add ip addr to bridge, %v", err)
		return err
	}
	return nil
}

func (m *BridgeManager) ReturnBridgeDetails() *BridgeManager {
	return m
}

func (m *BridgeManager) BringDownBridge() error {
	bridgeLink, err := netlink.LinkByName(m.DefaultBridge)
	if err != nil {
		log.Printf("could not find bridge %s: %v\n", m.DefaultBridge, err)
		return err
	}
	err = netlink.LinkSetDown(bridgeLink)
	if err != nil {
		log.Printf("could not bring bridge down %s: %v\n", m.DefaultBridge, err)
		return err
	}
	return nil
}

func (m *BridgeManager) BringUpBridge() error {
	bridgeLink, err := netlink.LinkByName(m.DefaultBridge)
	if err != nil {
		log.Printf("could not find bridge %s: %v\n", m.DefaultBridge, err)
		return err
	}
	err = netlink.LinkSetUp(bridgeLink)
	if err != nil {
		log.Printf("could not bring bridge up %s: %v\n", m.DefaultBridge, err)
		return err
	}
	return nil
}

func (m *BridgeManager) DeleteBridge() error {
	bridgeLink, err := netlink.LinkByName(m.DefaultBridge)
	if err != nil {
		log.Printf("could not find bridge %s: %v\n", m.DefaultBridge, err)
		return err
	}
	err = netlink.LinkDel(bridgeLink)
	if err != nil {
		log.Printf("could not delete bridge %s: %v\n", m.DefaultBridge, err)
		return err
	}
	return nil
}
