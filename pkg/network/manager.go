package network

import (
	"fmt"
	"net"
)

func NewNetworkManager() (*NetworkManager, error) {
	ipManager := &IPManager{
		Allocated: make(map[string]net.IP),
	}

	_, err := ipManager.InitIPManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize IP manager: %v", err)
	}

	bridgeManager := &BridgeManager{}
	err = bridgeManager.CreateBridgeWithIp(ipManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create bridge: %v", err)
	}

	vethManager := &VethManager{
		veths: make(map[string][2]string),
	}

	natManager := &NatManager{
		BridgeManager: bridgeManager,
		IpManager:     ipManager,
	}

	return &NetworkManager{
		BridgeManager: bridgeManager,
		IpManager:     ipManager,
		VethManager:   vethManager,
		NatManager:    natManager,
	}, nil
}
