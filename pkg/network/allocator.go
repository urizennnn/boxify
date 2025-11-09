package network

import (
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// func (m *IPManager) ReleaseIP(containerID string) error
// func (m *IPManager) GetGateway() string
// func (m *IPManager) AllocateIP(containerID string) (string, error)

func (m *IPManager) GetNextIP() string {
	if CheckNetworkConfigExists() {
		networkConfig, err := ReadNetworkConfig("default")
		if err != nil {
			log.Printf("GetNextIP: Failed to read config, falling back to memory: %v", err)
			return m.NextIP.String()
		}
		m.NextIP = net.ParseIP(networkConfig.Ipam.NextIP)
		return networkConfig.Ipam.NextIP
	}
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

	if CheckNetworkConfigExists() {
		log.Println("InitIPManager: Network config already exists, loading from file")
		networkConfig, err := ReadNetworkConfig("default")
		if err != nil {
			log.Printf("InitIPManager: Error reading existing config: %v", err)
			return "", err
		}

		m.BridgeCIDR = networkConfig.Ipam.Subnet
		m.Gateway = net.ParseIP(networkConfig.Ipam.Gateway)
		m.NextIP = net.ParseIP(networkConfig.Ipam.NextIP)
		m.Allocated = make(map[string]net.IP)

		for name, ipStr := range networkConfig.Ipam.AllocatedIPs {
			m.Allocated[name] = net.ParseIP(ipStr)
		}

		log.Printf("InitIPManager: Loaded existing network - Gateway: %s, NextIP: %s", m.Gateway, m.NextIP)
		return m.Gateway.String() + m.BridgeCIDR, nil
	}

	log.Println("InitIPManager: No existing config found, initializing new network")
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
			networkIP, _, _ := net.ParseCIDR(candidate)
			firstUsableIP := m.IncrementIp(networkIP.String())
			m.Gateway = firstUsableIP
			m.NextIP = firstUsableIP
			log.Printf("InitIPManager: Gateway: %v, NextIP: %v", m.Gateway, m.NextIP)
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

	if err := m.persistNextIP(result.String()); err != nil {
		log.Printf("IncrementIp: Warning - failed to persist NextIP: %v", err)
	}

	return result
}

func (m *IPManager) persistNextIP(nextIP string) error {
	if !CheckNetworkConfigExists() {
		return nil
	}

	configPath := NetworkStorageDir + "/default.yaml"
	lock := NewFileLock(configPath)
	if err := lock.AcquireLock(); err != nil {
		return err
	}
	defer lock.ReleaseLock()

	networkConfig, err := ReadNetworkConfig("default")
	if err != nil {
		return err
	}

	networkConfig.Ipam.NextIP = nextIP
	m.NextIP = net.ParseIP(nextIP)

	return WriteNetworkConfigWithoutLock(networkConfig)
}

func WriteNetworkConfigWithoutLock(networkStorage *NetworkStorage) error {
	configPath := NetworkStorageDir + "/default.yaml"

	data, err := yaml.Marshal(networkStorage)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
