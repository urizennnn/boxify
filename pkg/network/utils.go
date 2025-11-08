package network

import (
	"os"
	"strconv"

	"github.com/vishvananda/netlink"
)

func LinkExists(name string) (netlink.Link, error) {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return nil, err
	}
	return link, nil
}

func GetNsFD(pid int) (*os.File, error) {
	nsPath := "/proc/" + strconv.Itoa(pid) + "/ns/net"
	nsFile, err := os.Open(nsPath)
	if err != nil {
		return nil, err
	}
	return nsFile, nil
}

func GetOriginalNS() (*os.File, error) {
	originalNs, err := os.Open("/proc/self/ns/net")
	if err != nil {
		return nil, err
	}
	return originalNs, nil
}
