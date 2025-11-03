package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/urizennnn/boxify/pkg/cgroup"
	"github.com/urizennnn/boxify/pkg/container"
	"github.com/urizennnn/boxify/pkg/daemon/requests"
	"github.com/urizennnn/boxify/pkg/daemon/types"
	"github.com/urizennnn/boxify/pkg/network"
)

// DaemonInterface defines the methods needed from the daemon
type DaemonInterface interface {
	AddContainer(container *types.Container)
	GetContainer(id string) (*types.Container, error)
	NetworkManager() *network.NetworkManager
}

func HandleCreate(d DaemonInterface, w http.ResponseWriter, r *http.Request) {
	var request requests.InitContainerRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	memory := request.MemoryLimit
	cpu := request.CpuLimit

	containerID := uuid.New().String()

	parent(d, containerID, memory, cpu)
	switch os.Args[1] {
	case "child":
		child(d, os.Args[2], os.Args[3], os.Args[4], os.Args[5], os.Args[6], os.Args[7])
	}
}

func parent(d DaemonInterface, containerID, memory, cpu string) {
	networkMgr := d.NetworkManager()
	hostVeth, containerVeth, err := networkMgr.VethManager.CreateVethPairAndAttachToHostBridge(containerID, networkMgr.BridgeManager)
	if err != nil {
		log.Printf("Error creating veth pair: %v\n", err)
		return
	}

	gateway := networkMgr.IpManager.GetGateway()
	nextIP := networkMgr.IpManager.GetNextIP()

	cmd := exec.Command("/proc/self/exe", "child", containerID, memory, cpu, containerVeth, gateway, nextIP)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWNS,

		Unshareflags: syscall.CLONE_NEWNS,
	}
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	pid := cmd.Process.Pid

	containerInfo := &types.Container{
		ID:      containerID,
		PID:     pid,
		Image:   "",
		Command: []string{"/bin/sh"},
		NetworkInfo: &types.NetworkInfo{
			IP:            nextIP,
			Gateway:       gateway,
			Bridge:        networkMgr.BridgeManager.ReturnBridgeDetails().DefaultBridge,
			HostVeth:      hostVeth,
			ContainerVeth: containerVeth,
		},
		CreatedAt: time.Now(),
		Status:    "created",
	}

	d.AddContainer(containerInfo)

	if err := networkMgr.MoveVethIntoContainerNamespace(containerVeth, containerID, d); err != nil {
		log.Printf("Error moving veth into namespace: %v\n", err)
		return
	}

	err = cgroup.SetupCgroupsV2(pid, memory, cpu)
	if err != nil {
		log.Printf("Error setting up cgroups: %v\n", err)
		return
	}

	if err := saveContainerToJSON(containerInfo); err != nil {
		log.Printf("Error saving container info: %v\n", err)
	}

	err = cmd.Wait()
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}
}

func child(d DaemonInterface, containerID, mem, cpu, containerVeth, gateway, ipAddr string) {
	err, newContainerID := container.InitContainer()

	mergedDir := "/tmp/boxify-container/" + newContainerID + "/merged"
	defer syscall.Unmount(mergedDir, syscall.MNT_DETACH)
	defer os.RemoveAll("/tmp/boxify-container/" + newContainerID)
	defer os.RemoveAll("/sys/fs/cgroup/boxify/")

	if err != nil {
		log.Printf("Error: failed in creating overlay FS %v\n", err)
		os.Exit(1)
	}
	setupMounts()

	if err := d.NetworkManager().SetupContainerInterface(containerID, d); err != nil {
		log.Printf("Error setting up network: %v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command("/bin/sh")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Env = []string{"PATH=/bin:/usr/bin:/sbin:/usr/sbin"}
	if err := cmd.Run(); err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	log.Println("Container exiting...")
}

func saveContainerToJSON(container *types.Container) error {
	dir := "/var/lib/boxify"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return errors.New("failed to create directory: " + err.Error())
	}

	filePath := filepath.Join(dir, container.ID+".json")
	data, err := json.MarshalIndent(container, "", "  ")
	if err != nil {
		return errors.New("failed to marshal container info: " + err.Error())
	}

	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return errors.New("failed to write container info: " + err.Error())
	}

	return nil
}

func setupMounts() {
	log.Printf("setting up proc mount\n")
	err := syscall.Mount("proc", "/proc", "proc", 0, "")
	if err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	log.Printf("setting up sys mount\n")
	err = syscall.Mount("sysfs", "/sys", "sysfs", 0, "")
	if err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	log.Printf("setting up dev mount\n")
	err = syscall.Mount("tmpfs", "/dev", "tmpfs", 0, "")
	if err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
