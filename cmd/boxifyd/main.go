package main

import "net"

type Daemon struct {
	networkMgr *network.Manager
	containers map[string]*Container 
	listener   net.Listener
	stopChan   chan struct{}
}

func (d *Daemon) Start() error {
	// Initialize network
	d.networkMgr.Initialize()

	// Create Unix socket
	listener, err := net.Listen("unix", "/var/run/boxify.sock")

	// Accept connections
	for {
		conn, err := listener.Accept()
		go d.handleConnection(conn)
	}
}
