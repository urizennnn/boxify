package legacy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"syscall"

	"gopkg.in/yaml.v3"

	"github.com/urizennnn/boxify/config"
	"github.com/urizennnn/boxify/pkg/daemon/requests"
)

type httpResult struct {
	PID int    `json:"pid"`
	Cmd string `json:"cmd"`
}

// Run executes the bare container creation logic
func Run() {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", "/var/run/boxify.sock")
			},
		},
	}
	requestedConfig, err := parseConfigFile()
	if err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
		return
	}
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get current working directory: %v", err)
		return
	}
	reqBody := requests.InitContainerRequest{
		Name:         requestedConfig.ImageName,
		OriginFolder: cwd,
		MemoryLimit:  requestedConfig.ResourceLimits.MemoryLimit,
		CpuLimit:     requestedConfig.ResourceLimits.CpuLimit,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatalf("Failed to marshal request: %v", err)
	}
	resp, err := client.Post(
		"http://unix/containers/create",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Fatalf("Request failed with status: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
	}
	var result httpResult
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		log.Printf("Failed to unmarshal response: %v", err)
	}
	fmt.Printf("Container created successfully: PID=%d, Cmd=%s\n", result.PID, result.Cmd)

	containerEnv := []string{
		"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		"TERM=xterm",
		"HOME=/root",
		"HOSTNAME=container",
	}

	err = syscall.Exec(
		"/usr/bin/nsenter",
		[]string{
			"nsenter",
			"-t", fmt.Sprintf("%d", result.PID),
			"-u", "-i", "-p", "-n", "-m",
			"/bin/sh",
		},
		containerEnv,
	)
	if err != nil {
		fmt.Printf("Failed to exec nsenter: %v", err)
	}
}

func parseConfigFile() (config.ContainerConfig, error) {
	var yamlFile []byte
	yamlFile, err := os.ReadFile("boxify.yaml")
	if err != nil {
		yamlFile, err = os.ReadFile("boxify.yml")
		if err != nil {
			log.Fatalf("Config file not found: %v", err)
			return config.ContainerConfig{}, err
		}
		log.Fatalf("Failed to open config file: %v", err)
		return config.ContainerConfig{}, err
	}
	var fileConfig config.ContainerConfig
	err = yaml.Unmarshal(yamlFile, &fileConfig)
	if err != nil {
		log.Fatalf("Error unmarshaling YAML: %v", err)
		return config.ContainerConfig{}, err
	}

	fmt.Printf("Parsed config: %+v\n", fileConfig)
	return fileConfig, nil
}
