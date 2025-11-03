package network

import (
	"github.com/vishvananda/netlink"
)

func LinkExists(name string) (netlink.Link, error) {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return nil, err
	}
	return link, nil
}
