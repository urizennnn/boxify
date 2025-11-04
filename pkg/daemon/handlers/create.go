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
	"github.com/urizennnn/boxify/pkg/daemon/requests"
	"github.com/urizennnn/boxify/pkg/daemon/types"
	"github.com/urizennnn/boxify/pkg/network"
)

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

	cmd := exec.Command("/usr/local/bin/boxify-init", containerID, memory, cpu, containerVeth, gateway, nextIP)
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
