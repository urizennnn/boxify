package network

// TODO: implement port forwarding rules at a later time

import (
	"log"
	"os/exec"
)

func (m *NatManager) enableIPForwarding() error {
	cmd := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1")

	if err := cmd.Run(); err != nil {
		log.Printf("error enabling IP forwarding: %v (requires root/CAP_NET_ADMIN)", err)
		return nil
	}
	log.Printf("IP forwarding enabled successfully")
	return nil
}

// TODO: switch away from masquerading to SNAT with specific IP
func (m *NatManager) setupMasquerading() error {
	bridgeDetails := m.BridgeManager.ReturnBridgeDetails()
	ipCidr := m.IpManager.GetIpDetails()
	fullCIDR := ipCidr.Gateway.String() + ipCidr.BridgeCIDR

	checkCmd := exec.Command("iptables", "-t", "nat", "-C", "POSTROUTING", "-s", fullCIDR, "!", "-o", bridgeDetails.DefaultBridge, "-j", "MASQUERADE")
	if err := checkCmd.Run(); err == nil {
		log.Printf("Masquerading rule already exists for %s", fullCIDR)
		return nil
	}

	log.Printf("Setting up masquerading for network %s via bridge %s", fullCIDR, bridgeDetails.DefaultBridge)
	cmd := exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", fullCIDR, "!", "-o", bridgeDetails.DefaultBridge, "-j", "MASQUERADE")

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

	checkCmd := exec.Command("iptables", "-C", "FORWARD", "-i", bridgeDetails.DefaultBridge, "-j", "ACCEPT")
	if err := checkCmd.Run(); err == nil {
		log.Printf("Forwarding rules already exist for %s", bridgeDetails.DefaultBridge)
		return nil
	}

	log.Printf("Setting up forwarding rules for bridge %s", bridgeDetails.DefaultBridge)

	cmd1 := exec.Command("iptables", "-A", "FORWARD", "-i", bridgeDetails.DefaultBridge, "-j", "ACCEPT")
	if err := cmd1.Run(); err != nil {
		log.Printf("error setting up outbound forwarding rule: %v", err)
		return nil
	}

	cmd2 := exec.Command("iptables", "-A", "FORWARD", "-o", bridgeDetails.DefaultBridge, "-m", "conntrack", "--ctstate", "RELATED,ESTABLISHED", "-j", "ACCEPT")
	if err := cmd2.Run(); err != nil {
		log.Printf("error setting up inbound forwarding rule: %v", err)
		return nil
	}

	log.Printf("Forwarding rules set up successfully")
	return nil
}

func (m *NatManager) RemoveMasquerading() error {
	bridgeDetails := m.BridgeManager.ReturnBridgeDetails()
	ipCidr := m.IpManager.GetIpDetails()
	fullCIDR := ipCidr.Gateway.String() + ipCidr.BridgeCIDR

	cmd := exec.Command("iptables", "-t", "nat", "-D", "POSTROUTING", "-s", fullCIDR, "!", "-o", bridgeDetails.DefaultBridge, "-j", "MASQUERADE")
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
