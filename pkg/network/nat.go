package network

// TODO: implement port forwarding rules at a later time

import (
	"log"
	"os/exec"
)

func (m *NatManager) enableIPForwarding() error {
	cmd := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1")

	if err := cmd.Run(); err != nil {
		log.Printf("error enabling IP forwarding: %v", err)
		return nil
	}
	cmd = exec.Command("sysctl", "-p")
	if err := cmd.Run(); err != nil {
		log.Printf("error enabling IP forwarding: %v", err)
		return nil
	}

	return nil
}

// TODO: switch away from masquerading to SNAT with specific IP
func (m *NatManager) setupMasquerading() error {
	bridgeDetails := m.BridgeManager.ReturnBridgeDetails()
	ipCidr := m.IpManager.GetIpDetails()
	cmd := exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", ipCidr.BridgeCIDR, "!", "-o", bridgeDetails.DefaultBridge, "-j", "MASQUERADE")

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("error setting up masquerading: %v, output: %s", err, out)
		return nil
	}
	log.Printf("Masquerading setup output: %s\n", out)
	return nil
}

func (m *NatManager) SetupForwardingRules() error {
	bridgeDetails := m.BridgeManager.ReturnBridgeDetails()

	cmd := exec.Command("iptables", "-A", "FORWARD", "-i", "eth0", "-o", bridgeDetails.DefaultBridge, "-j", "ACCEPT")
	if err := cmd.Run(); err != nil {
		log.Printf("error setting up forwarding rules: %v", err)
		return nil
	}
	return nil
}

func (m *NatManager) RemoveMasquerading() error {
	bridgeDetails := m.BridgeManager.ReturnBridgeDetails()

	cmd := exec.Command("iptables", "-t", "nat", "-D", "POSTROUTING", "-s", m.IpManager.GetIpDetails().BridgeCIDR, "!", "-o", bridgeDetails.DefaultBridge, "-j", "MASQUERADE")
	if err := cmd.Run(); err != nil {
		log.Printf("Error removing masquerading: %v", err)
		return nil
	}
	return nil
}

func (m *NatManager) EnableNat() error {
	if err := m.enableIPForwarding(); err != nil {
		log.Printf("Error enabling IP forwarding: %v", err)
		return nil
	}
	if err := m.setupMasquerading(); err != nil {
		log.Printf("Error setting up masquerading: %v", err)
		return nil
	}
	if err := m.SetupForwardingRules(); err != nil {
		log.Printf("Error setting up forwarding rules: %v", err)
		return nil
	}
	return nil
}
