package network

import (
	"log"
	"net"

	"github.com/urizennnn/boxify/pkg/daemon/types"
	"github.com/vishvananda/netlink"
)

// ContainerGetter interface for getting container information
type ContainerGetter interface {
	GetContainer(id string) (*types.Container, error)
}

func (m *NetworkManager) MoveVethIntoContainerNamespace(vethName string, containerId string, damon ContainerGetter) error {
	container, err := damon.GetContainer(containerId)
	if err != nil {
		log.Printf("could not get container %v: %v", containerId, err)
		return nil
	}
	containerVeth, err := netlink.LinkByName(container.NetworkInfo.ContainerVeth)
	if err != nil {
		log.Printf("could not find veth %s: %v", container.NetworkInfo.ContainerVeth, err)
		return nil
	}

	err = netlink.LinkSetNsPid(
		containerVeth, container.PID,
	)
	if err != nil {
		log.Printf("could not move veth %s into namespace %d: %v\n", vethName, container.PID, err)
		return err
	}
	return nil
}

func (m *NetworkManager) RenameVethInContainerNamespace(newName string, containerId string, damon ContainerGetter) error {
	container, err := damon.GetContainer(containerId)
	if err != nil {
		log.Printf("could not get container %v: %v", containerId, err)
		return nil
	}

	containerVeth, err := netlink.LinkByName(container.NetworkInfo.ContainerVeth)
	if err != nil {
		log.Printf("could not find veth %s: %v", container.NetworkInfo.ContainerVeth, err)
		return nil
	}

	err = netlink.LinkSetName(containerVeth, newName)
	if err != nil {
		log.Printf("could not rename veth %s to %s: %v", container.NetworkInfo.ContainerVeth, newName, err)
		return nil
	}

	return nil
}

func (m *NetworkManager) AssignIPToVethInContainerNamespace(containerId string, damon ContainerGetter) error {
	container, err := damon.GetContainer(containerId)
	if err != nil {
		log.Printf("could not get container %v: %v", containerId, err)
		return nil
	}

	containerVeth, err := netlink.LinkByName(container.NetworkInfo.ContainerVeth)
	if err != nil {
		log.Printf("could not find veth %s: %v", container.NetworkInfo.ContainerVeth, err)
		return nil
	}

	ipAddr := m.IpManager.GetNextIP()
	if ipAddr == "" {
		log.Printf("container %v does not have an IP address", containerId)
		return nil
	}
	addr, err := netlink.ParseAddr(ipAddr)
	if err != nil {
		log.Printf("failed to parse addr, %v\n", err)
		return err
	}

	err = netlink.AddrAdd(containerVeth, addr)
	if err != nil {
		log.Printf("could not assign IP address %s to veth %s: %v", ipAddr, container.NetworkInfo.ContainerVeth, err)
		return nil
	}
	err = netlink.LinkSetUp(containerVeth)
	if err != nil {
		log.Printf("could not set link up, %v\n", err)
		return err
	}

	gateway := container.NetworkInfo.Gateway
	gw := net.ParseIP(gateway)
	_, dstNet, err := net.ParseCIDR("0.0.0.0/0")
	if err != nil {
		log.Printf("failed to parse default route CIDR: %v", err)
		return nil
	}

	route := &netlink.Route{
		Dst:       dstNet,
		Gw:        gw,
		LinkIndex: containerVeth.Attrs().Index,
	}
	if err = netlink.RouteAdd(route); err != nil {
		log.Printf("could not add route, %v\n", err)
		return err
	}
	return nil
}


func (m *NetworkManager) CreateLoopbackInContainerNamespace(containerId string, damon ContainerGetter) error {
	la := netlink.NewLinkAttrs()
	la.Name = "lo"

	existingLink, err := LinkExists(la.Name)
	if err == nil {
		err = netlink.LinkSetUp(existingLink)
		if err != nil {
			log.Printf("could not set up loopback interface: %v", err)
			return nil
		}
	} else {
		loopback := netlink.Device{LinkAttrs: la}
		err := netlink.LinkAdd(&loopback)
		if err != nil {
			log.Printf("could not add loopback interface: %v", err)
			return nil
		}
		err = netlink.LinkSetUp(&loopback)
		if err != nil {
			log.Printf("could not set up loopback interface: %v", err)
			return nil
		}
	}
	return nil
}

func SetupContainerNetworkStandalone(containerID, containerVethName, gateway, ipAddr string) error {
	containerVeth, err := netlink.LinkByName(containerVethName)
	if err != nil {
		log.Printf("could not find veth %s: %v", containerVethName, err)
		return err
	}

	err = netlink.LinkSetName(containerVeth, "eth0")
	if err != nil {
		log.Printf("could not rename veth %s to eth0: %v", containerVethName, err)
		return err
	}

	containerVeth, err = netlink.LinkByName("eth0")
	if err != nil {
		log.Printf("could not find renamed veth eth0: %v", err)
		return err
	}

	addr, err := netlink.ParseAddr(ipAddr)
	if err != nil {
		log.Printf("failed to parse addr %s: %v\n", ipAddr, err)
		return err
	}

	err = netlink.AddrAdd(containerVeth, addr)
	if err != nil {
		log.Printf("could not assign IP address %s to eth0: %v", ipAddr, err)
		return err
	}

	err = netlink.LinkSetUp(containerVeth)
	if err != nil {
		log.Printf("could not set link up: %v\n", err)
		return err
	}

	gw := net.ParseIP(gateway)
	_, dstNet, err := net.ParseCIDR("0.0.0.0/0")
	if err != nil {
		log.Printf("failed to parse default route CIDR: %v", err)
		return err
	}

	route := &netlink.Route{
		Dst:       dstNet,
		Gw:        gw,
		LinkIndex: containerVeth.Attrs().Index,
	}
	if err = netlink.RouteAdd(route); err != nil {
		log.Printf("could not add route: %v\n", err)
		return err
	}

	la := netlink.NewLinkAttrs()
	la.Name = "lo"

	existingLink, err := LinkExists(la.Name)
	if err == nil {
		err = netlink.LinkSetUp(existingLink)
		if err != nil {
			log.Printf("could not set up loopback interface: %v", err)
			return err
		}
	} else {
		loopback := netlink.Device{LinkAttrs: la}
		err := netlink.LinkAdd(&loopback)
		if err != nil {
			log.Printf("could not add loopback interface: %v", err)
			return err
		}
		err = netlink.LinkSetUp(&loopback)
		if err != nil {
			log.Printf("could not set up loopback interface: %v", err)
			return err
		}
	}

	return nil
}

func (m *NetworkManager) SetupContainerInterface(containerId string, damon ContainerGetter) error {
	err := m.RenameVethInContainerNamespace("eth0", containerId, damon)
	if err != nil {
		log.Printf("failed to rename veth to eth0: %v", err)
		return nil
	}

	err = m.AssignIPToVethInContainerNamespace(containerId, damon)
	if err != nil {
		log.Printf("failed to assign IP and configure interface: %v", err)
		return nil
	}

	err = m.CreateLoopbackInContainerNamespace(containerId, damon)
	if err != nil {
		log.Printf("failed to setup loopback interface: %v", err)
		return nil
	}
	if err = m.NatManager.EnableNat(); err != nil {
		log.Printf("failed to enable NAT: %v", err)
		return nil
	}

	return nil
}
