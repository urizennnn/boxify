package network

import (
	"log"
	"net"

	"github.com/urizennnn/boxify/pkg/daemon/types"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

// ContainerGetter interface for getting container information
type ContainerGetter interface {
	GetContainer(id string) (*types.Container, error)
}

func (m *NetworkManager) MoveVethIntoContainerNamespace(vethName string, containerId string, damon ContainerGetter) error {
	log.Printf("[MoveVeth] Starting move of veth %s for container %s", vethName, containerId)

	container, err := damon.GetContainer(containerId)
	if err != nil {
		log.Printf("[MoveVeth] Could not get container %v: %v", containerId, err)
		return nil
	}
	log.Printf("[MoveVeth] Container PID: %d, ContainerVeth: %s", container.PID, container.NetworkInfo.ContainerVeth)

	containerVeth, err := netlink.LinkByName(container.NetworkInfo.ContainerVeth)
	if err != nil {
		log.Printf("[MoveVeth] Could not find veth %s: %v", container.NetworkInfo.ContainerVeth, err)
		return nil
	}
	log.Printf("[MoveVeth] Found veth %s (index: %d)", container.NetworkInfo.ContainerVeth, containerVeth.Attrs().Index)

	containerFD, err := GetNsFD(container.PID)
	if err != nil {
		log.Printf("[MoveVeth] Could not get netns fd for container %s: %v", containerId, err)
		return nil
	}
	log.Printf("[MoveVeth] Got container namespace FD: %d for PID %d", containerFD.Fd(), container.PID)

	log.Printf("[MoveVeth] Moving veth %s into namespace FD %d", vethName, containerFD.Fd())
	err = netlink.LinkSetNsFd(
		containerVeth, int(containerFD.Fd()),
	)
	if err != nil {
		log.Printf("[MoveVeth] Could not move veth %s into namespace %d: %v", vethName, container.PID, err)
		return err
	}
	log.Printf("[MoveVeth] Successfully moved veth into namespace")

	log.Printf("[MoveVeth] Switching to container namespace (FD: %d)", containerFD.Fd())
	if err := netns.Set(netns.NsHandle(containerFD.Fd())); err != nil {
		log.Printf("[MoveVeth] Failed to switch to container namespace: %v", err)
		return err
	}
	log.Printf("[MoveVeth] Successfully switched to container namespace")

	return nil
}

func (m *NetworkManager) RenameVethInContainerNamespace(newName string, containerId string, damon ContainerGetter) error {
	log.Printf("[RenameVeth] Starting rename to %s for container %s", newName, containerId)

	container, err := damon.GetContainer(containerId)
	if err != nil {
		log.Printf("[RenameVeth] Could not get container %v: %v", containerId, err)
		return nil
	}
	log.Printf("[RenameVeth] Looking for veth: %s", container.NetworkInfo.ContainerVeth)

	containerVeth, err := netlink.LinkByName(container.NetworkInfo.ContainerVeth)
	if err != nil {
		log.Printf("[RenameVeth] Could not find veth %s: %v", container.NetworkInfo.ContainerVeth, err)
		return nil
	}
	log.Printf("[RenameVeth] Found veth %s (index: %d)", container.NetworkInfo.ContainerVeth, containerVeth.Attrs().Index)

	log.Printf("[RenameVeth] Renaming veth %s to %s", container.NetworkInfo.ContainerVeth, newName)
	err = netlink.LinkSetName(containerVeth, newName)
	if err != nil {
		log.Printf("[RenameVeth] Could not rename veth %s to %s: %v", container.NetworkInfo.ContainerVeth, newName, err)
		return nil
	}
	log.Printf("[RenameVeth] Successfully renamed veth to %s", newName)

	return nil
}

func (m *NetworkManager) AssignIPToVethInContainerNamespace(containerId string, damon ContainerGetter) error {
	log.Printf("[AssignIP] Starting IP assignment for container %s", containerId)

	container, err := damon.GetContainer(containerId)
	if err != nil {
		log.Printf("[AssignIP] Could not get container %v: %v", containerId, err)
		return nil
	}
	log.Printf("[AssignIP] Looking for veth: %s","eth0")

	containerVeth, err := netlink.LinkByName("eth0")
	if err != nil {
		log.Printf("[AssignIP] Could not find veth %s: %v", "eth0", err)
		return nil
	}
	log.Printf("[AssignIP] Found veth %s (index: %d)", "eth0", containerVeth.Attrs().Index)

	ipAddr := m.IpManager.GetNextIP()
	if ipAddr == "" {
		log.Printf("[AssignIP] Container %v does not have an IP address", containerId)
		return nil
	}
	log.Printf("[AssignIP] Got IP address: %s", ipAddr)

	//FIX: dynamically set subnet mask
	addr, err := netlink.ParseAddr(ipAddr+"/16")
	if err != nil {
		log.Printf("[AssignIP] Failed to parse addr %s: %v", ipAddr, err)
		return err
	}

	log.Printf("[AssignIP] Adding IP address %s to veth", ipAddr)
	err = netlink.AddrAdd(containerVeth, addr)
	if err != nil {
		log.Printf("[AssignIP] Could not assign IP address %s to veth %s: %v", ipAddr, "eth0", err)
		return nil
	}
	log.Printf("[AssignIP] Successfully added IP address")

	log.Printf("[AssignIP] Setting link up")
	err = netlink.LinkSetUp(containerVeth)
	if err != nil {
		log.Printf("[AssignIP] Could not set link up: %v", err)
		return err
	}
	log.Printf("[AssignIP] Link is up")

	gateway := container.NetworkInfo.Gateway
	log.Printf("[AssignIP] Adding default route via gateway %s", gateway)
	gw := net.ParseIP(gateway)
	_, dstNet, err := net.ParseCIDR("0.0.0.0/0")
	if err != nil {
		log.Printf("[AssignIP] Failed to parse default route CIDR: %v", err)
		return nil
	}

	route := &netlink.Route{
		Dst:       dstNet,
		Gw:        gw,
		LinkIndex: containerVeth.Attrs().Index,
	}
	if err = netlink.RouteAdd(route); err != nil {
		log.Printf("[AssignIP] Could not add route: %v", err)
		return err
	}
	log.Printf("[AssignIP] Successfully added default route")

	return nil
}

func (m *NetworkManager) CreateLoopbackInContainerNamespace(containerId string, damon ContainerGetter) error {
	log.Printf("[Loopback] Starting loopback setup for container %s", containerId)

	la := netlink.NewLinkAttrs()
	la.Name = "lo"

	log.Printf("[Loopback] Checking if loopback interface 'lo' exists")
	existingLink, err := LinkExists(la.Name)
	if err == nil {
		log.Printf("[Loopback] Loopback interface exists, setting it up")
		err = netlink.LinkSetUp(existingLink)
		if err != nil {
			log.Printf("[Loopback] Could not set up existing loopback interface: %v", err)
			return nil
		}
		log.Printf("[Loopback] Successfully set up existing loopback interface")
	} else {
		log.Printf("[Loopback] Loopback interface does not exist, creating new one")
		loopback := netlink.Device{LinkAttrs: la}
		err := netlink.LinkAdd(&loopback)
		if err != nil {
			log.Printf("[Loopback] Could not add loopback interface: %v", err)
			return nil
		}
		log.Printf("[Loopback] Created loopback interface, setting it up")
		err = netlink.LinkSetUp(&loopback)
		if err != nil {
			log.Printf("[Loopback] Could not set up new loopback interface: %v", err)
			return nil
		}
		log.Printf("[Loopback] Successfully created and set up loopback interface")
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

func (m *NetworkManager) SetupContainerInterface(containerId string, damon ContainerGetter, containerVeth string) error {
	log.Printf("[SetupInterface] ========== Starting network setup for container %s ==========", containerId)

	log.Printf("[SetupInterface] Saving original namespace")
	origNS, err := GetOriginalNS()
	if err != nil {
		log.Printf("[SetupInterface] Failed to get original namespace: %v", err)
		return err
	}
	log.Printf("[SetupInterface] Original namespace saved (FD: %d)", origNS.Fd())

	defer func() {
		log.Printf("[SetupInterface] Restoring original namespace (FD: %d)", origNS.Fd())
		if err := netns.Set(netns.NsHandle(origNS.Fd())); err != nil {
			log.Printf("[SetupInterface] FAILED to restore original namespace: %v", err)
		} else {
			log.Printf("[SetupInterface] Successfully restored to original namespace")
		}
	}()

	log.Printf("[SetupInterface] Step 1: Moving veth %s into container namespace", containerVeth)
	if err := m.MoveVethIntoContainerNamespace(containerVeth, containerId, damon); err != nil {
		log.Printf("[SetupInterface] Error moving veth into namespace: %v", err)
		return err
	}
	log.Printf("[SetupInterface] Step 1 completed")

	log.Printf("[SetupInterface] Step 2: Renaming veth to eth0")
	err = m.RenameVethInContainerNamespace("eth0", containerId, damon)
	if err != nil {
		log.Printf("[SetupInterface] Failed to rename veth to eth0: %v", err)
		return nil
	}
	log.Printf("[SetupInterface] Step 2 completed")

	log.Printf("[SetupInterface] Step 3: Assigning IP and configuring interface")
	err = m.AssignIPToVethInContainerNamespace(containerId, damon)
	if err != nil {
		log.Printf("[SetupInterface] Failed to assign IP and configure interface: %v", err)
		return nil
	}
	log.Printf("[SetupInterface] Step 3 completed")

	log.Printf("[SetupInterface] Step 4: Setting up loopback interface")
	err = m.CreateLoopbackInContainerNamespace(containerId, damon)
	if err != nil {
		log.Printf("[SetupInterface] Failed to setup loopback interface: %v", err)
		return nil
	}
	log.Printf("[SetupInterface] Step 4 completed")

	// Restore to host namespace before enabling NAT (NAT rules must be on host)
	log.Printf("[SetupInterface] Restoring to host namespace before NAT setup")
	if err := netns.Set(netns.NsHandle(origNS.Fd())); err != nil {
		log.Printf("[SetupInterface] Failed to restore to host namespace: %v", err)
		return err
	}
	log.Printf("[SetupInterface] Successfully restored to host namespace")

	log.Printf("[SetupInterface] Step 5: Enabling NAT")
	if err = m.NatManager.EnableNat(); err != nil {
		log.Printf("[SetupInterface] Failed to enable NAT: %v", err)
		return nil
	}
	log.Printf("[SetupInterface] Step 5 completed")

	log.Printf("[SetupInterface] ========== Network setup completed for container %s ==========", containerId)
	return nil
}
