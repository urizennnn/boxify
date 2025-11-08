package network

import (
	"log"

	"github.com/vishvananda/netlink"
)

func (m *VethManager) CreateVethPairAndAttachToHostBridge(containerID string, bridge *BridgeManager) (string, string, error) {
	hostName := "veth-" + containerID[:8]
	containerName := "vethc-" + containerID[:8]

	existingLink, err := LinkExists(hostName)
	var veth *netlink.Veth

	if err == nil {
		existingVeth, ok := existingLink.(*netlink.Veth)
		if ok {
			veth = existingVeth
		} else {
			veth = &netlink.Veth{
				LinkAttrs: netlink.LinkAttrs{
					Name: hostName,
				},
				PeerName: containerName,
			}
		}
	} else {
		veth = &netlink.Veth{
			LinkAttrs: netlink.LinkAttrs{
				Name: hostName,
			},
			PeerName: containerName,
		}
		if err := netlink.LinkAdd(veth); err != nil {
			return "", "", err
		}
	}
	bridgeManagerDetails := bridge.ReturnBridgeDetails()

	err = netlink.LinkSetMaster(veth, bridgeManagerDetails.BridgeInstance)
	if err != nil {
		log.Printf("could not set link master, %v\n", err)
		return "", "", err
	}

	err = netlink.LinkSetUp(veth)
	if err != nil {
		log.Printf("could not set link up, %v\n", err)
		return "", "", err
	}
	log.Printf("veth name %v\n", veth.PeerName)
	m.veths[containerID] = [2]string{hostName, containerName}
	return hostName, containerName, nil
}

func (m *VethManager) DeleteVethPair(containerID string) error {
	vethNames, exists := m.veths[containerID]
	if !exists {
		log.Printf("veth pair not found for container ID: %s", containerID)
		return nil
	}

	for _, vethName := range vethNames {
		veth, err := netlink.LinkByName(vethName)
		if err != nil {
			log.Printf("could not find link by name, %v\n", err)
			return err
		}

		if err := netlink.LinkDel(veth); err != nil {
			log.Printf("could not delete link, %v\n", err)
			return err
		}
	}

	delete(m.veths, containerID)
	return nil
}
