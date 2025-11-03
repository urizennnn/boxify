package network
//TODO: implement port forwarding rules at a later time

import (
	"fmt"
	"os/exec"
)

func (m *NatManager) enableIPForwarding() error {
	cmd := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error enabling IP forwarding: %v", err)
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
		return fmt.Errorf("error setting up masquerading: %v, output: %s", err, out)
	}
	fmt.Printf("Masquerading setup output: %s\n", out)
	return nil
}

func (m *NatManager) SetupForwardingRules() error {
	bridgeDetails := m.BridgeManager.ReturnBridgeDetails()

	cmd := exec.Command("iptables", "-A", "FORWARD", "-i", bridgeDetails.DefaultBridge, "-o", bridgeDetails.DefaultBridge, "-j", "ACCEPT")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error setting up forwarding rules: %v", err)
	}
	return nil
}

func (m *NatManager) RemoveMasquerading() error {
	bridgeDetails := m.BridgeManager.ReturnBridgeDetails()

	cmd := exec.Command("iptables", "-t", "nat", "-D", "POSTROUTING", "-s", m.IpManager.GetIpDetails().BridgeCIDR, "!", "-o", bridgeDetails.DefaultBridge, "-j", "MASQUERADE")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Error removing masquerading: %v", err)
	}
	return nil
}

func (m *NatManager) EnableNat() error {
	if err := m.enableIPForwarding(); err != nil {
		return fmt.Errorf("Error enabling IP forwarding: %v", err)
	}
	if err := m.setupMasquerading(); err != nil {
		return fmt.Errorf("Error setting up masquerading: %v", err)
	}
	if err := m.SetupForwardingRules(); err != nil {
		return fmt.Errorf("Error setting up forwarding rules: %v", err)
	}
	return nil
}
