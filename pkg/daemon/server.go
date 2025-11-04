package daemon

import (
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
)

func (d *Daemon) Init() {
	setupLogging()

	if err := writePIDFile(); err != nil {
		log.Printf("Warning: couldn't write PID file: %v", err)
	}
	defer removePIDFile()

	if err := os.Remove("/var/run/boxify.sock"); err != nil && !os.IsNotExist(err) {
		log.Fatalf("Failed to remove old socket: %v", err)
	}

	listener, err := net.Listen("unix", "/var/run/boxify.sock")
	if err != nil {
		log.Fatalf("Failed to create socket: %v", err)
	}
	defer listener.Close()

	if err := os.Chmod("/var/run/boxify.sock", 0o660); err != nil {
		log.Fatalf("Failed to set permissions: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/containers/create", d.HandleCreateRequest)
	// mux.HandleFunc("/containers/start", handleStart)
	// mux.HandleFunc("/containers/stop", handleStop)
	// mux.HandleFunc("/containers/list", handleList)
	// mux.HandleFunc("/containers/remove", handleRemove)
	//
	log.Println("Boxify daemon started, listening on /var/run/boxify.sock")

	if err := http.Serve(listener, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func setupLogging() {
	if os.Getenv("INVOCATION_ID") != "" {
		log.SetOutput(os.Stdout)
	} else {
		logFile, err := os.OpenFile("/var/log/boxify.log",
			os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err == nil {
			log.SetOutput(logFile)
		}
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func writePIDFile() error {
	pid := os.Getpid()
	pidStr := strconv.Itoa(pid)

	return os.WriteFile("/var/run/boxify.pid",
		[]byte(pidStr), 0o644)
}

func removePIDFile() {
	os.Remove("/var/run/boxify.pid")
}
