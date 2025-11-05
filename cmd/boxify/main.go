package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"syscall"

	"gopkg.in/yaml.v3"

	"github.com/urizennnn/boxify/config"
	"github.com/urizennnn/boxify/pkg/daemon/requests"
)

func main() {
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
		MemoryLimit:  requestedConfig.Settings.MemoryLimit,
		CpuLimit:     requestedConfig.Settings.CpuLimit,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatalf("Failed to marshal request: %v", err)
	}
	file, err := os.OpenFile("/proc/self/ns/net", 0, 0o600)
	if err != nil {
		fmt.Printf("Failed to open net ns: %v", err)
		return
	}
	err = os.WriteFile("/var/lib/boxify/networks/host", []byte{byte(file.Fd())}, 0o644)
	if err != nil {
		fmt.Printf("Failed to write host net ns: %v", err)
		return
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

	var result int
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Failed to decode response: %v", err)
	}

	fmt.Printf("Container created successfully: %+v\n", result)
	err = syscall.Exec(
		"/usr/bin/nsenter",
		[]string{
			"nsenter",
			"-t", fmt.Sprintf("%d", result),
			"-u", "-i", "-p", "-n", "-m",
			"/bin/sh",
		},
		os.Environ(),
	)
	if err != nil {
		fmt.Printf("Failed to exec nsenter: %v", err)
	}
}

func parseConfigFile() (config.ConfigStructure, error) {
	yamlFile, err := os.ReadFile("boxify.yaml")
	if err != nil {
		yamlFile, err = os.ReadFile("boxify.yml")
		if err != nil {
			log.Fatalf("Config file not found: %v", err)
			return config.ConfigStructure{}, err
		}
		log.Fatalf("Failed to open config file: %v", err)
		return config.ConfigStructure{}, err
	}
	var fileConfig config.ConfigStructure
	err = yaml.Unmarshal(yamlFile, &fileConfig)
	if err != nil {
		log.Fatalf("Error unmarshaling YAML: %v", err)
		return config.ConfigStructure{}, err
	}

	fmt.Printf("Parsed config: %+v\n", fileConfig)
	return fileConfig, nil
}
