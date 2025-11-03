package network

import (
	"fmt"

	"github.com/vishvananda/netlink"
)

func (m *VethManager) CreateVethPairAndAttachToHostBridge(containerID string, bridge *BridgeManager) (string, string, error) {
	hostName := "veth-" + containerID[:8]
	containerName := "vethc-" + containerID[:8]

	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: hostName,
		},
		PeerName: containerName,
	}

	if err := netlink.LinkAdd(veth); err != nil {
		return "", "", err
	}
	bridgeManagerDetails := bridge.ReturnBridgeDetails()

	err := netlink.LinkSetMaster(veth, bridgeManagerDetails.bridgeInstance)
	if err != nil {
		fmt.Printf("could not set link master, %v\n", err)
		return "", "", err
	}

	err = netlink.LinkSetUp(veth)
	if err != nil {
		fmt.Printf("could not set link up, %v\n", err)
		return "", "", err
	}
	m.veths[containerID] = [2]string{hostName, containerName}
	return hostName, containerName, nil
}

func (m *VethManager) DeleteVethPair(containerID string) error {
	vethNames, exists := m.veths[containerID]
	if !exists {
		return fmt.Errorf("veth pair not found for container ID: %s", containerID)
	}

	for _, vethName := range vethNames {
		veth, err := netlink.LinkByName(vethName)
		if err != nil {
			fmt.Printf("could not find link by name, %v\n", err)
			return err
		}

		if err := netlink.LinkDel(veth); err != nil {
			fmt.Printf("could not delete link, %v\n", err)
			return err
		}
	}

	delete(m.veths, containerID)
	return nil
}
