package network

import (
	"log"
	"net"
)

func NewNetworkManager() (*NetworkManager, error) {
	ipManager := &IPManager{
		Allocated: make(map[string]net.IP),
	}

	_, err := ipManager.InitIPManager()
	if err != nil {
		log.Printf("failed to initialize IP manager: %v", err)
		return nil, nil
	}

	bridgeManager := &BridgeManager{}
	err = bridgeManager.CreateBridgeWithIp(ipManager)
	if err != nil {
		log.Printf("failed to create bridge: %v", err)
		return nil, nil
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
