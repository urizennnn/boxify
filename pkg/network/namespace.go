package network

import (
	"fmt"
	"net"

	"github.com/urizennnn/boxify/pkg/daemon"
	"github.com/vishvananda/netlink"
)

func (m *NetworkManager) MoveVethIntoContainerNamespace(vethName string, containerId string, damon *daemon.Daemon) error {
	container, err := damon.GetContainer(containerId)
	if err != nil {
		return fmt.Errorf("could not get container %v: %v", containerId, err)
	}
	containerVeth, err := netlink.LinkByName(container.NetworkInfo.ContainerVeth)
	if err != nil {
		return fmt.Errorf("could not find veth %s: %v", container.NetworkInfo.ContainerVeth, err)
	}

	err = netlink.LinkSetNsPid(
		containerVeth, container.PID,
	)
	if err != nil {
		fmt.Printf("could not move veth %s into namespace %d: %v\n", vethName, container.PID, err)
		return err
	}
	return nil
}

func (m *NetworkManager) RenameVethInContainerNamespace(newName string, containerId string, damon *daemon.Daemon) error {
	container, err := damon.GetContainer(containerId)
	if err != nil {
		return fmt.Errorf("could not get container %v: %v", containerId, err)
	}

	containerVeth, err := netlink.LinkByName(container.NetworkInfo.ContainerVeth)
	if err != nil {
		return fmt.Errorf("could not find veth %s: %v", container.NetworkInfo.ContainerVeth, err)
	}

	err = netlink.LinkSetName(containerVeth, newName)
	if err != nil {
		return fmt.Errorf("could not rename veth %s to %s: %v", container.NetworkInfo.ContainerVeth, newName, err)
	}

	return nil
}

func (m *NetworkManager) AssignIPToVethInContainerNamespace(containerId string, damon *daemon.Daemon) error {
	container, err := damon.GetContainer(containerId)
	if err != nil {
		return fmt.Errorf("could not get container %v: %v", containerId, err)
	}

	containerVeth, err := netlink.LinkByName(container.NetworkInfo.ContainerVeth)
	if err != nil {
		return fmt.Errorf("could not find veth %s: %v", container.NetworkInfo.ContainerVeth, err)
	}

	ipAddr := m.ipManager.GetNextIP()
	if ipAddr == "" {
		return fmt.Errorf("container %v does not have an IP address", containerId)
	}
	addr, err := netlink.ParseAddr(ipAddr)
	if err != nil {
		fmt.Printf("failed to parse addr, %v\n", err)
		return err
	}

	err = netlink.AddrAdd(containerVeth, addr)
	if err != nil {
		return fmt.Errorf("could not assign IP address %s to veth %s: %v", ipAddr, container.NetworkInfo.ContainerVeth, err)
	}
	err = netlink.LinkSetUp(containerVeth)
	if err != nil {
		fmt.Printf("could not set link up, %v\n", err)
		return err
	}
	dst := &net.IPNet{
		IP:   net.IPv4zero,
		Mask: net.CIDRMask(0, 32),
	}
	gw := net.ParseIP(ipAddr)

	route := &netlink.Route{
		Dst:       dst,
		Gw:        gw,
		LinkIndex: containerVeth.Attrs().Index,
	}
	if err = netlink.RouteAdd(route); err != nil {
		fmt.Printf("could not add route, %v\n", err)
		return err
	}
	return nil
}


func (m *NetworkManager) CreateLoopbackInContainerNamespace(containerId string, damon *daemon.Daemon) error {
	la := netlink.NewLinkAttrs()
	la.Name = "lo"

	existingLink, err := LinkExists(la.Name)
	if err == nil {
		err = netlink.LinkSetUp(existingLink)
		if err != nil {
			return fmt.Errorf("could not set up loopback interface: %v", err)
		}
	} else {
		loopback := netlink.Device{LinkAttrs: la}
		err := netlink.LinkAdd(&loopback)
		if err != nil {
			return fmt.Errorf("could not add loopback interface: %v", err)
		}
		err = netlink.LinkSetUp(&loopback)
		if err != nil {
			return fmt.Errorf("could not set up loopback interface: %v", err)
		}
	}
	return nil
}

func (m *NetworkManager) SetupContainerInterface(containerId string, damon *daemon.Daemon) error {
	err := m.RenameVethInContainerNamespace("eth0", containerId, damon)
	if err != nil {
		return fmt.Errorf("failed to rename veth to eth0: %v", err)
	}

	err = m.AssignIPToVethInContainerNamespace(containerId, damon)
	if err != nil {
		return fmt.Errorf("failed to assign IP and configure interface: %v", err)
	}

	err = m.CreateLoopbackInContainerNamespace(containerId, damon)
	if err != nil {
		return fmt.Errorf("failed to setup loopback interface: %v", err)
	}
	if err = m.natManager.EnableNat(); err != nil {
		return fmt.Errorf("failed to enable NAT: %v", err)
	}

	return nil
}
