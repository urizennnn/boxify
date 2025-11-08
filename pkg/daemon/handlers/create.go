package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/urizennnn/boxify/config"
	"github.com/urizennnn/boxify/pkg/cgroup"
	"github.com/urizennnn/boxify/pkg/container"
	"github.com/urizennnn/boxify/pkg/daemon/requests"
	"github.com/urizennnn/boxify/pkg/daemon/types"
	"github.com/urizennnn/boxify/pkg/network"
	"gopkg.in/yaml.v3"
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
	networkMgr := d.NetworkManager()
	hostVeth, containerVeth, err := networkMgr.VethManager.CreateVethPairAndAttachToHostBridge(containerID, networkMgr.BridgeManager)
	log.Printf("Created veth pair: host=%s, container=%s\n", hostVeth, containerVeth)
	if err != nil {
		log.Printf("Error creating veth pair: %v\n", err)
		return
	}

	pid, err := parent(d, containerID, memory, cpu, containerVeth, hostVeth, networkMgr)
	if err != nil {
		http.Error(w, "Failed to create container", http.StatusInternalServerError)
		return
	}
	_, err = http.ResponseWriter.Write(w, []byte(strconv.Itoa(pid)))
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

func parent(d DaemonInterface, containerID, memory, cpu string, containerVeth, hostVeth string, networkMgr *network.NetworkManager) (int, error) {
	gateway := networkMgr.IpManager.GetGateway()
	bridgeCIDR := networkMgr.IpManager.BridgeCIDR
	nextIP := networkMgr.IpManager.GetNextIP() + bridgeCIDR
	err, mergedDir := container.InitContainer(containerID)
	if err != nil {
		log.Fatalf("Error: failed in creating overlay FS %v\n", err)
	}

	cmd := exec.Command("/usr/local/bin/boxify-init", containerID, memory, cpu, mergedDir)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWNS,

		Unshareflags: syscall.CLONE_NEWNS,
	}
	// Don't inherit daemon's stdin/stdout/stderr - container will be accessed via nsenter
	// Keep stderr for init logs, but don't set stdin/stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Printf("Error starting container: %v\n", err)
		return 0, err
	}
	pid := cmd.Process.Pid

	// Start a goroutine to wait for the container process and reap it when it exits
	go func() {
		if err := cmd.Wait(); err != nil {
			log.Printf("Container %s (PID %d) exited with error: %v", containerID, pid, err)
		} else {
			log.Printf("Container %s (PID %d) exited successfully", containerID, pid)
		}
	}()

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

	log.Printf("Setting up container interface for container %s\n", containerID)
	if err := networkMgr.SetupContainerInterface(containerID, d, containerVeth); err != nil {
		log.Printf("Error setting up container interface: %v\n", err)
		return 0, err
	}

	err = cgroup.SetupCgroupsV2(pid, memory, cpu)
	if err != nil {
		log.Printf("Error setting up cgroups: %v\n", err)
		return 0, err
	}

	if err = saveContainerToDefaultConfig(containerInfo); err != nil {
		log.Printf("Error saving container info: %v\n", err)
	}

	return pid, nil
}

func saveContainerToDefaultConfig(container *types.Container) error {
	filename := "/var/lib/boxify/networks/default.yaml"
	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		log.Printf("Error reading default network config: %v\n", err)
		return err
	}

	var configStructure config.NetworkStorage
	if err := yaml.Unmarshal(yamlFile, &configStructure); err != nil {
		log.Printf("Error unmarshaling YAML: %v\n", err)
		return err
	}

	configStructure.Containers = append(configStructure.Containers, container)
	updatedData, err := yaml.Marshal(&configStructure)
	if err != nil {
		log.Printf("Error marshaling updated YAML: %v\n", err)
		return err
	}

	if err := os.WriteFile(filename, updatedData, 0o644); err != nil {
		return errors.New("failed to write container info: " + err.Error())
	}

	return nil
}
