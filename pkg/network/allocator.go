package network

import (
	"fmt"
	"net"
	"strings"
)

// func (m *IPManager) ReleaseIP(containerID string) error
// func (m *IPManager) GetGateway() string
// func (m *IPManager) AllocateIP(containerID string) (string, error)

func (m *IPManager) GetNextIP() string {
	return m.nextIP.String()
}

func (m *IPManager) GetGateway() string {
	return m.gateway.String()
}


func (m *IPManager) GetHostNetworks() ([]*net.IPNet, error) {
	var networks []*net.IPNet

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		// Skip loopback
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipnet.IP.To4() != nil { // IPv4 only
					networks = append(networks, ipnet)
				}
			}
		}
	}

	return networks, nil
}

func isNetworkConflict(cidr string, existing []*net.IPNet) bool {
	_, testNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return true
	}

	for _, existingNet := range existing {
		if existingNet.Contains(testNet.IP) || testNet.Contains(existingNet.IP) {
			return true
		}
	}

	return false
}

func (m *IPManager) InitIPManager() (string, error) {
	hostNetworks, err := m.GetHostNetworks()
	if err != nil {
		return "", err
	}

	candidates := []string{
		"172.17.0.0/16",
		"172.18.0.0/16",
		"10.88.0.0/16",
		"192.168.100.0/16",
	}

	for _, candidate := range candidates {
		if !isNetworkConflict(candidate, hostNetworks) {
			m.bridgeCIDR = "/16"
			m.gateway, _, _ = net.ParseCIDR(candidate)
			m.nextIP = m.IncrementIp(m.gateway.String())
		}
	}

	return "", fmt.Errorf("could not find non-conflicting network")
}

func (m *IPManager) IncrementIp(ip string) net.IP {
	splitString := strings.Split(ip, ".")
	for i := len(splitString) - 1; i >= 0; i-- {
		if splitString[i] != "255" {
			splitString[i] = fmt.Sprintf("%d", net.ParseIP(splitString[i]).To4()[i]+1)
			break
		}
	}
	return net.ParseIP(strings.Join(splitString, "."))
}
