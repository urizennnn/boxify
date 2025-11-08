package container

import (
	"log"
	"os"

	"github.com/urizennnn/boxify/config"
	"github.com/urizennnn/boxify/pkg/daemon/types"
	"gopkg.in/yaml.v3"
)

func ListAllContainers() []*types.Container {
	yamlConfigFile, err := os.ReadFile("/var/lib/boxify/networks/default.yaml")
	if err != nil {
		log.Printf("Error opening default network config: %v\n", err)
		return nil
	}
	var fileConfig config.NetworkStorage
	err = yaml.Unmarshal(yamlConfigFile, &fileConfig)
	if err != nil {
		log.Printf("Error unmarshaling YAML: %v\n", err)
		return nil
	}
	return fileConfig.Containers
}
