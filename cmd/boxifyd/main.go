package main

import (
"net"
	"github.com/urizennnn/boxify/pkg/daemon"
)

func (d *daemon.Daemon) Start() error {
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
