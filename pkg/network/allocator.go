package network

import (
	"log"
	"net"
	"strconv"
	"strings"
)

// func (m *IPManager) ReleaseIP(containerID string) error
// func (m *IPManager) GetGateway() string
// func (m *IPManager) AllocateIP(containerID string) (string, error)

func (m *IPManager) GetNextIP() string {
	return m.NextIP.String()
}

func (m *IPManager) GetIpDetails() *IPManager {
	return m
}

func (m *IPManager) GetGateway() string {
	return m.Gateway.String()
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
	log.Println("InitIPManager: Starting")
	hostNetworks, err := m.GetHostNetworks()
	if err != nil {
		log.Printf("InitIPManager: Error getting host networks: %v", err)
		return "", err
	}
	log.Printf("InitIPManager: Found %d host networks", len(hostNetworks))

	candidates := []string{
		"172.17.0.0/16",
		"172.18.0.0/16",
		"10.88.0.0/16",
		"192.168.100.0/16",
	}

	for _, candidate := range candidates {
		log.Printf("InitIPManager: Checking candidate: %s", candidate)
		if !isNetworkConflict(candidate, hostNetworks) {
			log.Printf("InitIPManager: Using network: %s", candidate)
			m.BridgeCIDR = "/16"
			m.Gateway, _, _ = net.ParseCIDR(candidate)
			log.Printf("InitIPManager: Gateway: %v", m.Gateway)
			m.NextIP = m.IncrementIp(m.Gateway.String())
			return candidate, nil
		}
	}

	log.Println("InitIPManager: could not find non-conflicting network")
	return "", nil
}

func (m *IPManager) IncrementIp(ip string) net.IP {
	log.Printf("IncrementIp: Input IP: %s", ip)
	splitString := strings.Split(ip, ".")
	log.Printf("IncrementIp: Split IP into %d parts: %v", len(splitString), splitString)
	for i := len(splitString) - 1; i >= 0; i-- {
		if splitString[i] != "255" {
			log.Printf("IncrementIp: Incrementing octet at index %d: %s", i, splitString[i])
			transformedInt, err := strconv.Atoi(splitString[i])
			if err != nil {
				log.Printf("IncrementIp: Error converting octet to int: %v", err)
				return nil
			}

			splitString[i] = strconv.Itoa(transformedInt + 1)
			break
		}
	}
	result := net.ParseIP(strings.Join(splitString, "."))
	log.Printf("IncrementIp: Result: %v", result)
	return result
}
